package restaurant

// import (
// 	"context"
// 	"errors"
// 	"testing"

// 	"delivery-service/internal/domain"
// )

// type mockRestaurantRepo struct {
// 	getByIDFunc    func(ctx context.Context, id int64) (*domain.Restaurant, error)
// 	listActiveFunc func(ctx context.Context) ([]*domain.Restaurant, error)
// }

// func (m *mockRestaurantRepo) GetByID(ctx context.Context, id int64) (*domain.Restaurant, error) {
// 	return m.getByIDFunc(ctx, id)
// }
// func (m *mockRestaurantRepo) ListActive(ctx context.Context) ([]*domain.Restaurant, error) {
// 	return m.listActiveFunc(ctx)
// }

// type mockMenuRepo struct {
// 	getMenuFunc func(ctx context.Context, restaurantID int64) ([]*domain.MenuItem, error)
// }

// func (m *mockMenuRepo) GetMenu(ctx context.Context, restaurantID int64) ([]*domain.MenuItem, error) {
// 	return m.getMenuFunc(ctx, restaurantID)
// }

// func TestService_GetMenu(t *testing.T) {
// 	ctx := context.Background()

// 	t.Run("success get menu for active restaurant", func(t *testing.T) {
// 		restRepo := &mockRestaurantRepo{
// 			getByIDFunc: func(ctx context.Context, id int64) (*domain.Restaurant, error) {
// 				return &domain.Restaurant{ID: id, IsActive: true}, nil
// 			},
// 		}
// 		menuRepo := &mockMenuRepo{
// 			getMenuFunc: func(ctx context.Context, restaurantID int64) ([]*domain.MenuItem, error) {
// 				return []*domain.MenuItem{
// 					{ID: 1, Name: "Борщ", Price: 300, IsAvailable: true},
// 				}, nil
// 			},
// 		}

// 		svc := NewService(restRepo, menuRepo)
// 		items, err := svc.GetMenu(ctx, 1)

// 		if err != nil {
// 			t.Fatalf("expected no error, got: %v", err)
// 		}
// 		if len(items) != 1 || items[0].Name != "Борщ" {
// 			t.Errorf("unexpected menu items returned")
// 		}
// 	})

// 	t.Run("error when restaurant is inactive", func(t *testing.T) {
// 		restRepo := &mockRestaurantRepo{
// 			getByIDFunc: func(ctx context.Context, id int64) (*domain.Restaurant, error) {
// 				return &domain.Restaurant{ID: id, IsActive: false}, nil // Ресторан временно закрыт
// 			},
// 		}
// 		menuRepo := &mockMenuRepo{}

// 		svc := NewService(restRepo, menuRepo)
// 		_, err := svc.GetMenu(ctx, 1)

// 		if !errors.Is(err, domain.ErrRestaurantInactive) {
// 			t.Errorf("expected ErrRestaurantInactive, got: %v", err)
// 		}
// 	})
// }
