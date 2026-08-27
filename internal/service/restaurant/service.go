package restaurant

import (
	"context"
	"fmt"

	"delivery-service/internal/domain"
)

type RestaurantRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Restaurant, error)
	ListActive(ctx context.Context) ([]*domain.Restaurant, error)
}

type MenuRepository interface {
	GetMenu(ctx context.Context, restaurantID int64) ([]*domain.MenuItem, error)
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

func (s *Service) ListActiveRestaurants(ctx context.Context) ([]*domain.Restaurant, error) {
	restaurants, err := s.restaurantRepo.ListActive(ctx)
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
