package restaurant

import (
	"context"
	"fmt"

	"delivery-service/internal/domain"
)

type RestaurantRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Restaurant, error)
	ListActive(ctx context.Context, estType domain.EstablishmentType) ([]*domain.Restaurant, error)
	Create(ctx context.Context, rest *domain.Restaurant) error
}

type MenuRepository interface {
	GetMenu(ctx context.Context, restaurantID int64) ([]*domain.MenuItem, error)
	Create(ctx context.Context, item *domain.MenuItem) error
}

type Service struct {
	restaurantRepo RestaurantRepository
	menuRepo       MenuRepository
}

func NewService(restaurantRepo RestaurantRepository, menuRepo MenuRepository) *Service {
	return &Service{
		restaurantRepo: restaurantRepo,
		menuRepo:       menuRepo,
	}
}

func (s *Service) CreateRestaurant(ctx context.Context, dto domain.CreateRestaurantDTO) (*domain.Restaurant, error) {
	rest := domain.Restaurant{
		Name:       dto.Name,
		Type:       dto.Type,
		Address:    dto.Address,
		WebhookURL: dto.WebhookURL,
		IsActive:   true,
	}

	if err := s.restaurantRepo.Create(ctx, &rest); err != nil {
		return nil, fmt.Errorf("restaurant service - create: %w", err)
	}

	return &rest, nil
}

func (s *Service) ListActiveRestaurants(ctx context.Context, estType domain.EstablishmentType) ([]*domain.Restaurant, error) {
	restaurants, err := s.restaurantRepo.ListActive(ctx, estType)
	if err != nil {
		return nil, fmt.Errorf("restaurant service - list active: %w", err)
	}
	return restaurants, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*domain.Restaurant, error) {
	rest, err := s.restaurantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("restaurant service - get by id: %w", err)
	}
	return rest, nil
}

func (s *Service) GetMenu(ctx context.Context, restaurantID int64) ([]*domain.MenuItem, error) {

	rest, err := s.restaurantRepo.GetByID(ctx, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("restaurant service - check restaurant: %w", err)
	}
	if !rest.IsActive {
		return nil, domain.ErrRestaurantInactive
	}

	items, err := s.menuRepo.GetMenu(ctx, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("restaurant service - get menu: %w", err)
	}
	return items, nil
}

func (s *Service) CreateMenu(ctx context.Context, restID int64, dto domain.CreateMenuItemDTO) (*domain.MenuItem, error) {
	menuItem := domain.MenuItem{
		RestaurantID:  restID,
		Name:          dto.Name,
		Description:   dto.Description,
		Price:         dto.Price,
		StockQuantity: dto.StockQuantity,
		IsAvailable:   true,
	}

	if err := s.menuRepo.Create(ctx, &menuItem); err != nil {
		return nil, fmt.Errorf("restaurant service - create menu %w", err)
	}

	return &menuItem, nil
}
