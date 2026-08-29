package restaurant

import (
	"context"
	"errors"
	"testing"

	"delivery-service/internal/domain"
)

type mockRestaurantRepo struct {
	getByIDFunc    func(ctx context.Context, id int64) (*domain.Restaurant, error)
	listActiveFunc func(ctx context.Context, estType domain.EstablishmentType) ([]*domain.Restaurant, error)
	createFunc     func(ctx context.Context, rest *domain.Restaurant) error
}

func (m *mockRestaurantRepo) GetByID(ctx context.Context, id int64) (*domain.Restaurant, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockRestaurantRepo) ListActive(ctx context.Context, estType domain.EstablishmentType) ([]*domain.Restaurant, error) {
	if m.listActiveFunc != nil {
		return m.listActiveFunc(ctx, estType)
	}
	return nil, nil
}

func (m *mockRestaurantRepo) Create(ctx context.Context, rest *domain.Restaurant) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, rest)
	}
	return nil
}

type mockMenuRepo struct {
	getMenuFunc func(ctx context.Context, restaurantID int64) ([]*domain.MenuItem, error)
	createFunc  func(ctx context.Context, item *domain.MenuItem) error
}

func (m *mockMenuRepo) GetMenu(ctx context.Context, restaurantID int64) ([]*domain.MenuItem, error) {
	if m.getMenuFunc != nil {
		return m.getMenuFunc(ctx, restaurantID)
	}
	return nil, nil
}

func (m *mockMenuRepo) Create(ctx context.Context, item *domain.MenuItem) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, item)
	}
	return nil
}

func TestService_GetMenu(t *testing.T) {
	ctx := context.Background()

	t.Run("success get menu for active restaurant", func(t *testing.T) {
		restRepo := &mockRestaurantRepo{
			getByIDFunc: func(ctx context.Context, id int64) (*domain.Restaurant, error) {
				return &domain.Restaurant{ID: id, IsActive: true}, nil
			},
		}
		menuRepo := &mockMenuRepo{
			getMenuFunc: func(ctx context.Context, restaurantID int64) ([]*domain.MenuItem, error) {
				return []*domain.MenuItem{
					{ID: 1, Name: "Борщ", Price: 300, IsAvailable: true},
				}, nil
			},
		}

		svc := NewService(restRepo, menuRepo)
		items, err := svc.GetMenu(ctx, 1)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(items) != 1 || items[0].Name != "Борщ" {
			t.Errorf("unexpected menu items returned")
		}
	})

	t.Run("error when restaurant is inactive", func(t *testing.T) {
		restRepo := &mockRestaurantRepo{
			getByIDFunc: func(ctx context.Context, id int64) (*domain.Restaurant, error) {
				return &domain.Restaurant{ID: id, IsActive: false}, nil
			},
		}
		menuRepo := &mockMenuRepo{}

		svc := NewService(restRepo, menuRepo)
		_, err := svc.GetMenu(ctx, 1)

		if !errors.Is(err, domain.ErrRestaurantInactive) {
			t.Errorf("expected ErrRestaurantInactive, got: %v", err)
		}
	})

	t.Run("error when restaurant repository fails", func(t *testing.T) {
		expectedErr := errors.New("db connection failed")
		restRepo := &mockRestaurantRepo{
			getByIDFunc: func(ctx context.Context, id int64) (*domain.Restaurant, error) {
				return nil, expectedErr
			},
		}
		menuRepo := &mockMenuRepo{}

		svc := NewService(restRepo, menuRepo)
		_, err := svc.GetMenu(ctx, 1)

		if err == nil || !errors.Is(err, expectedErr) {
			t.Fatalf("expected db error wrapped, got: %v", err)
		}
	})
}

func TestService_CreateRestaurant(t *testing.T) {
	ctx := context.Background()

	t.Run("success create restaurant", func(t *testing.T) {
		restRepo := &mockRestaurantRepo{
			createFunc: func(ctx context.Context, rest *domain.Restaurant) error {
				rest.ID = 10
				return nil
			},
		}

		svc := NewService(restRepo, nil)
		dto := domain.CreateRestaurantDTO{
			Name:       "Додо Пицца",
			Type:       domain.EstablishmentTypeRestaurant,
			Address:    "Невский пр., 1",
			WebhookURL: "http://webhook.local",
		}

		res, err := svc.CreateRestaurant(ctx, dto)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if res.ID != 10 || res.Name != dto.Name || !res.IsActive {
			t.Errorf("unexpected restaurant created: %+v", res)
		}
	})
}

func TestService_CreateMenu(t *testing.T) {
	ctx := context.Background()

	t.Run("success create menu item", func(t *testing.T) {
		menuRepo := &mockMenuRepo{
			createFunc: func(ctx context.Context, item *domain.MenuItem) error {
				item.ID = 5
				return nil
			},
		}

		svc := NewService(nil, menuRepo)
		dto := domain.CreateMenuItemDTO{
			Name:          "Пицца Пепперони",
			Description:   "Острая пицца",
			Price:         550,
			StockQuantity: 10,
		}

		item, err := svc.CreateMenu(ctx, 1, dto)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if item.ID != 5 || item.RestaurantID != 1 || !item.IsAvailable {
			t.Errorf("unexpected menu item created: %+v", item)
		}
	})
}
