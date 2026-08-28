package order

import (
	"context"
	"errors"
	"fmt"

	"delivery-service/internal/domain"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}

type RestaurantRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Restaurant, error)
}

type MenuRepository interface {
	GetMenuItemsByIDs(ctx context.Context, ids []int64) ([]*domain.MenuItem, error)
}

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order, deductStock bool) error
	GetByID(ctx context.Context, id int64) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus) error
	ListByUserID(ctx context.Context, userID int64) ([]*domain.Order, error)
}

type Service struct {
	userRepo       UserRepository
	restaurantRepo RestaurantRepository
	menuRepo       MenuRepository
	orderRepo      OrderRepository
	orderEvents    chan<- *domain.Order
}

func NewService(
	userRepo UserRepository,
	restaurantRepo RestaurantRepository,
	menuRepo MenuRepository,
	orderRepo OrderRepository,
	orderEvents chan<- *domain.Order,
) *Service {
	return &Service{
		userRepo:       userRepo,
		restaurantRepo: restaurantRepo,
		menuRepo:       menuRepo,
		orderRepo:      orderRepo,
		orderEvents:    orderEvents,
	}
}

func (s *Service) CreateOrder(ctx context.Context, dto domain.CreateOrderDTO) (*domain.Order, error) {
	if len(dto.Items) == 0 {
		return nil, errors.New("order must contain at least one item")
	}

	user, err := s.userRepo.GetByID(ctx, dto.UserID)
	if err != nil {
		return nil, fmt.Errorf("order service - check user: %w", err)
	}

	restaurant, err := s.restaurantRepo.GetByID(ctx, dto.RestaurantID)
	if err != nil {
		return nil, fmt.Errorf("order service - check restaurant: %w", err)
	}
	if !restaurant.IsActive {
		return nil, domain.ErrRestaurantInactive
	}

	itemIDs := make([]int64, 0, len(dto.Items))
	for _, item := range dto.Items {
		if item.Quantity <= 0 {
			return nil, errors.New("item quantity must be greater than zero")
		}
		itemIDs = append(itemIDs, item.MenuItemID)
	}

	dbMenuItems, err := s.menuRepo.GetMenuItemsByIDs(ctx, itemIDs)
	if err != nil {
		return nil, fmt.Errorf("order service - get menu items: %w", err)
	}

	menuMap := make(map[int64]*domain.MenuItem, len(dbMenuItems))
	for _, item := range dbMenuItems {
		menuMap[item.ID] = item
	}

	isShop := restaurant.Type == domain.EstablishmentTypeShop
	var totalPrice int64
	orderItems := make([]domain.OrderItem, 0, len(dto.Items))

	for _, reqItem := range dto.Items {
		dbItem, exists := menuMap[reqItem.MenuItemID]
		if !exists {
			return nil, fmt.Errorf("item with id %d not found", reqItem.MenuItemID)
		}
		if dbItem.RestaurantID != dto.RestaurantID {
			return nil, fmt.Errorf("item %d does not belong to restaurant %d", dbItem.ID, dto.RestaurantID)
		}
		if !dbItem.IsAvailable {
			return nil, fmt.Errorf("item '%s' is currently unavailable", dbItem.Name)
		}

		// Проверка остатка на складе магазина
		if isShop && dbItem.StockQuantity < reqItem.Quantity {
			return nil, fmt.Errorf("not enough stock for '%s': available %d, requested %d", dbItem.Name, dbItem.StockQuantity, reqItem.Quantity)
		}

		totalPrice += dbItem.Price * int64(reqItem.Quantity)
		orderItems = append(orderItems, domain.OrderItem{
			MenuItemID: dbItem.ID,
			Quantity:   reqItem.Quantity,
			Price:      dbItem.Price,
		})
	}

	deliveryAddr := dto.DeliveryAddress
	if deliveryAddr == "" {
		deliveryAddr = user.Address
	}

	order := &domain.Order{
		UserID:          dto.UserID,
		RestaurantID:    dto.RestaurantID,
		Status:          domain.OrderStatusCreated,
		DeliveryAddress: deliveryAddr,
		TotalPrice:      totalPrice,
		Items:           orderItems,
	}

	// Вызываем атомарное создание со списанием остатков
	if err := s.orderRepo.Create(ctx, order, isShop); err != nil {
		return nil, fmt.Errorf("order service - save order: %w", err)
	}

	if s.orderEvents != nil {
		select {
		case s.orderEvents <- order:
		default:
		}
	}

	return order, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

func (s *Service) ListUserOrders(ctx context.Context, userID int64) ([]*domain.Order, error) {
	return s.orderRepo.ListByUserID(ctx, userID)
}

func (s *Service) UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus) error {
	return s.orderRepo.UpdateStatus(ctx, id, status)
}
