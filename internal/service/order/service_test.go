package order

import (
	"context"
	"errors"
	"testing"

	"delivery-service/internal/domain"
)

// --- Mocks ---

type mockUserRepo struct {
	getByIDFunc func(ctx context.Context, id int64) (*domain.User, error)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

type mockRestaurantRepo struct {
	getByIDFunc func(ctx context.Context, id int64) (*domain.Restaurant, error)
}

func (m *mockRestaurantRepo) GetByID(ctx context.Context, id int64) (*domain.Restaurant, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

type mockMenuRepo struct {
	getMenuItemsByIDsFunc func(ctx context.Context, ids []int64) ([]*domain.MenuItem, error)
}

func (m *mockMenuRepo) GetMenuItemsByIDs(ctx context.Context, ids []int64) ([]*domain.MenuItem, error) {
	if m.getMenuItemsByIDsFunc != nil {
		return m.getMenuItemsByIDsFunc(ctx, ids)
	}
	return nil, nil
}

type mockOrderRepo struct {
	createFunc       func(ctx context.Context, order *domain.Order, deductStock bool) error
	getByIDFunc      func(ctx context.Context, id int64) (*domain.Order, error)
	updateStatusFunc func(ctx context.Context, id int64, status domain.OrderStatus) error
	listByUserIDFunc func(ctx context.Context, userID int64) ([]*domain.Order, error)
}

func (m *mockOrderRepo) Create(ctx context.Context, order *domain.Order, deductStock bool) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, order, deductStock)
	}
	return nil
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, id, status)
	}
	return nil
}

func (m *mockOrderRepo) ListByUserID(ctx context.Context, userID int64) ([]*domain.Order, error) {
	if m.listByUserIDFunc != nil {
		return m.listByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

// --- Tests ---

func TestService_CreateOrder(t *testing.T) {
	ctx := context.Background()

	validUser := &domain.User{ID: 1, Name: "Ivan", Address: "Default Address 10"}
	activeRestaurant := &domain.Restaurant{
		ID:       10,
		Name:     "Pizza Place",
		Type:     domain.EstablishmentTypeRestaurant,
		IsActive: true,
	}

	t.Run("empty items validation error", func(t *testing.T) {
		svc := NewService(nil, nil, nil, nil, nil)
		_, err := svc.CreateOrder(ctx, domain.CreateOrderDTO{Items: []domain.CreateOrderItemDTO{}})
		if err == nil || err.Error() != "order must contain at least one item" {
			t.Fatalf("expected empty items error, got: %v", err)
		}
	})

	t.Run("inactive restaurant error", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
				return validUser, nil
			},
		}
		restRepo := &mockRestaurantRepo{
			getByIDFunc: func(ctx context.Context, id int64) (*domain.Restaurant, error) {
				return &domain.Restaurant{ID: 10, IsActive: false}, nil
			},
		}

		svc := NewService(userRepo, restRepo, nil, nil, nil)
		dto := domain.CreateOrderDTO{
			UserID:       1,
			RestaurantID: 10,
			Items:        []domain.CreateOrderItemDTO{{MenuItemID: 1, Quantity: 1}},
		}

		_, err := svc.CreateOrder(ctx, dto)
		if !errors.Is(err, domain.ErrRestaurantInactive) {
			t.Fatalf("expected ErrRestaurantInactive, got: %v", err)
		}
	})

	t.Run("item does not belong to restaurant error", func(t *testing.T) {
		userRepo := &mockUserRepo{getByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) { return validUser, nil }}
		restRepo := &mockRestaurantRepo{getByIDFunc: func(ctx context.Context, id int64) (*domain.Restaurant, error) { return activeRestaurant, nil }}
		menuRepo := &mockMenuRepo{
			getMenuItemsByIDsFunc: func(ctx context.Context, ids []int64) ([]*domain.MenuItem, error) {
				return []*domain.MenuItem{
					{ID: 1, RestaurantID: 999, Name: "Alien Pizza", Price: 500, IsAvailable: true},
				}, nil
			},
		}

		svc := NewService(userRepo, restRepo, menuRepo, nil, nil)
		dto := domain.CreateOrderDTO{
			UserID:       1,
			RestaurantID: 10,
			Items:        []domain.CreateOrderItemDTO{{MenuItemID: 1, Quantity: 1}},
		}

		_, err := svc.CreateOrder(ctx, dto)
		if err == nil {
			t.Fatalf("expected foreign item error, got nil")
		}
	})

	t.Run("shop not enough stock error", func(t *testing.T) {
		shopRest := &domain.Restaurant{
			ID:       20,
			Name:     "Corner Market",
			Type:     domain.EstablishmentTypeShop,
			IsActive: true,
		}
		userRepo := &mockUserRepo{getByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) { return validUser, nil }}
		restRepo := &mockRestaurantRepo{getByIDFunc: func(ctx context.Context, id int64) (*domain.Restaurant, error) { return shopRest, nil }}
		menuRepo := &mockMenuRepo{
			getMenuItemsByIDsFunc: func(ctx context.Context, ids []int64) ([]*domain.MenuItem, error) {
				return []*domain.MenuItem{
					{ID: 5, RestaurantID: 20, Name: "Milk", Price: 100, StockQuantity: 2, IsAvailable: true},
				}, nil
			},
		}

		svc := NewService(userRepo, restRepo, menuRepo, nil, nil)
		dto := domain.CreateOrderDTO{
			UserID:       1,
			RestaurantID: 20,
			Items:        []domain.CreateOrderItemDTO{{MenuItemID: 5, Quantity: 5}}, // Запросили 5, в наличии 2
		}

		_, err := svc.CreateOrder(ctx, dto)
		if err == nil {
			t.Fatalf("expected out of stock error, got nil")
		}
	})

	t.Run("success create order and default address", func(t *testing.T) {
		userRepo := &mockUserRepo{getByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) { return validUser, nil }}
		restRepo := &mockRestaurantRepo{getByIDFunc: func(ctx context.Context, id int64) (*domain.Restaurant, error) { return activeRestaurant, nil }}
		menuRepo := &mockMenuRepo{
			getMenuItemsByIDsFunc: func(ctx context.Context, ids []int64) ([]*domain.MenuItem, error) {
				return []*domain.MenuItem{
					{ID: 1, RestaurantID: 10, Name: "Pizza", Price: 500, IsAvailable: true},
				}, nil
			},
		}

		orderRepo := &mockOrderRepo{
			createFunc: func(ctx context.Context, order *domain.Order, deductStock bool) error {
				order.ID = 100
				return nil
			},
		}

		eventsChan := make(chan *domain.Order, 1)
		svc := NewService(userRepo, restRepo, menuRepo, orderRepo, eventsChan)

		dto := domain.CreateOrderDTO{
			UserID:          1,
			RestaurantID:    10,
			DeliveryAddress: "", // Пустой — должен подтянуться из пользователя
			Items:           []domain.CreateOrderItemDTO{{MenuItemID: 1, Quantity: 2}},
		}

		order, err := svc.CreateOrder(ctx, dto)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if order.ID != 100 {
			t.Errorf("expected assigned order ID 100, got %d", order.ID)
		}
		if order.TotalPrice != 1000 {
			t.Errorf("expected total price 1000, got %d", order.TotalPrice)
		}
		if order.DeliveryAddress != validUser.Address {
			t.Errorf("expected delivery address '%s', got '%s'", validUser.Address, order.DeliveryAddress)
		}

		// Проверяем отправку события в канал воркера
		select {
		case ev := <-eventsChan:
			if ev.ID != 100 {
				t.Errorf("expected event for order 100, got %d", ev.ID)
			}
		default:
			t.Errorf("expected order event to be sent to channel")
		}
	})
}
