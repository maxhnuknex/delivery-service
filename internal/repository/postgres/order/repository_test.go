package order

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"delivery-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type testFixtures struct {
	UserID       int64
	RestaurantID int64
	MenuItemID   int64
}

func setupTestDB(t *testing.T) (*pgxpool.Pool, testFixtures, func()) {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://test_user:test_password@localhost:5433/delivery-service_delivery-service_test?sslmode=disable"
	}
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	_, err = pool.Exec(ctx, "TRUNCATE TABLE order_items, orders, menu_items, restaurants, users RESTART IDENTITY CASCADE;")
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	var f testFixtures
	err = pool.QueryRow(ctx, "INSERT INTO users (name, phone, address) VALUES ('Тест', '+79998887766', 'Адрес') RETURNING id").Scan(&f.UserID)
	if err != nil {
		t.Fatalf("failed to create user fixture: %v", err)
	}

	err = pool.QueryRow(ctx, "INSERT INTO restaurants (name, address, is_active, webhook_url) VALUES ('Кафе', 'Улица', true, 'http://hook') RETURNING id").Scan(&f.RestaurantID)
	if err != nil {
		t.Fatalf("failed to create restaurant fixture: %v", err)
	}

	err = pool.QueryRow(ctx, "INSERT INTO menu_items (restaurant_id, name, price, is_available) VALUES ($1, 'Пицца', 500, true) RETURNING id", f.RestaurantID).Scan(&f.MenuItemID)
	if err != nil {
		t.Fatalf("failed to create menu item fixture: %v", err)
	}

	teardown := func() {
		_, _ = pool.Exec(ctx, "TRUNCATE TABLE order_items, orders, menu_items, restaurants, users RESTART IDENTITY CASCADE;")
		pool.Close()
	}

	return pool, f, teardown
}

func TestRepository(t *testing.T) {
	pool, f, teardown := setupTestDB(t)
	defer teardown()

	repo := NewRepository(pool)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("success create and get order with items", func(t *testing.T) {
		newOrder := &domain.Order{
			UserID:          f.UserID,
			RestaurantID:    f.RestaurantID,
			Status:          domain.OrderStatusCreated,
			DeliveryAddress: "ул. Ленина, 5",
			TotalPrice:      1000,
			Items: []domain.OrderItem{
				{MenuItemID: f.MenuItemID, Quantity: 2, Price: 500},
			},
		}

		err := repo.Create(ctx, newOrder, false)
		if err != nil {
			t.Fatalf("expected no error on order creation, got: %v", err)
		}
		if newOrder.ID == 0 {
			t.Fatalf("expected order ID to be set")
		}

		savedOrder, err := repo.GetByID(ctx, newOrder.ID)
		if err != nil {
			t.Fatalf("expected no error getting order, got: %v", err)
		}
		if savedOrder.TotalPrice != 1000 {
			t.Errorf("expected price 1000, got %d", savedOrder.TotalPrice)
		}
		if len(savedOrder.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(savedOrder.Items))
		}
		if savedOrder.Items[0].Quantity != 2 {
			t.Errorf("expected item quantity 2, got %d", savedOrder.Items[0].Quantity)
		}
	})

	t.Run("update order status", func(t *testing.T) {
		newOrder := &domain.Order{
			UserID:          f.UserID,
			RestaurantID:    f.RestaurantID,
			Status:          domain.OrderStatusCreated,
			DeliveryAddress: "ул. Ленина, 5",
			TotalPrice:      500,
			Items: []domain.OrderItem{
				{MenuItemID: f.MenuItemID, Quantity: 1, Price: 500},
			},
		}
		_ = repo.Create(ctx, newOrder, false)

		err := repo.UpdateStatus(ctx, newOrder.ID, domain.OrderStatusAccepted)
		if err != nil {
			t.Fatalf("expected no error on update status, got: %v", err)
		}

		updated, _ := repo.GetByID(ctx, newOrder.ID)
		if updated.Status != domain.OrderStatusAccepted {
			t.Errorf("expected status %s, got %s", domain.OrderStatusAccepted, updated.Status)
		}
	})

	t.Run("get by id not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 999999)
		if !errors.Is(err, domain.ErrOrderNotFound) {
			t.Errorf("expected ErrOrderNotFound, got: %v", err)
		}
	})
}
