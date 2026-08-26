package menu

import (
	"context"
	"strings"
	"testing"
	"time"

	"delivery-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, int64, func()) {
	t.Helper()

	dbURL := "postgres://delivery-service_user:delivery-service_password@localhost:5433/delivery-service_delivery-service_test?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	_, err = pool.Exec(ctx, "TRUNCATE TABLE menu_items, restaurants RESTART IDENTITY CASCADE;")
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	// Создаем тестовый ресторан для привязки блюд
	var restID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO restaurants (name, address, is_active, webhook_url)
		VALUES ('Test Bistro', 'Nevsky 10', true, 'http://mock.url')
		RETURNING id
	`).Scan(&restID)
	if err != nil {
		t.Fatalf("failed to create test restaurant: %v", err)
	}

	teardown := func() {
		_, _ = pool.Exec(ctx, "TRUNCATE TABLE menu_items, restaurants RESTART IDENTITY CASCADE;")
		pool.Close()
	}

	return pool, restID, teardown
}

func TestRepository(t *testing.T) {
	pool, restID, teardown := setupTestDB(t)
	defer teardown()

	repo := NewRepository(pool)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("success create and get menu with availability filter", func(t *testing.T) {
		item1 := &domain.MenuItem{
			RestaurantID: restID,
			Name:         "Бургер",
			Description:  "Сочная котлета",
			Price:        450,
			IsAvailable:  true,
		}
		item2 := &domain.MenuItem{
			RestaurantID: restID,
			Name:         "Картошка Фри (стоп-лист)",
			Description:  "Хрустящая",
			Price:        200,
			IsAvailable:  false, // В стоп-листе
		}

		if err := repo.Create(ctx, item1); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if err := repo.Create(ctx, item2); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		menu, err := repo.GetMenu(ctx, restID)
		if err != nil {
			t.Fatalf("expected no error getting menu, got: %v", err)
		}

		if len(menu) != 1 {
			t.Fatalf("expected 1 available item, got %d", len(menu))
		}
		if menu[0].Name != "Бургер" {
			t.Errorf("expected Бургер, got %s", menu[0].Name)
		}
	})

	t.Run("batch get by ids", func(t *testing.T) {
		item1 := &domain.MenuItem{RestaurantID: restID, Name: "Кола", Price: 150, IsAvailable: true}
		item2 := &domain.MenuItem{RestaurantID: restID, Name: "Чай", Price: 100, IsAvailable: true}
		_ = repo.Create(ctx, item1)
		_ = repo.Create(ctx, item2)

		items, err := repo.GetMenuItemsByIDs(ctx, []int64{item1.ID, item2.ID})
		if err != nil {
			t.Fatalf("expected no error on batch get, got: %v", err)
		}
		if len(items) != 2 {
			t.Errorf("expected 2 items, got %d", len(items))
		}
	})

	t.Run("foreign key violation", func(t *testing.T) {
		invalidItem := &domain.MenuItem{
			RestaurantID: 999999, // несуществующий ресторан
			Name:         "Пицца",
			Price:        600,
			IsAvailable:  true,
		}

		err := repo.Create(ctx, invalidItem)
		if err == nil {
			t.Fatalf("expected foreign key violation error, got nil")
		}
		if !strings.Contains(err.Error(), "SQLSTATE 23503") { // 23503 - FK error в Postgres
			t.Errorf("expected FK error code 23503, got: %v", err)
		}
	})
}
