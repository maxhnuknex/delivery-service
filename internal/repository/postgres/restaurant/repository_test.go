package restaurant

import (
	"context"
	"errors"
	"testing"
	"time"

	"delivery-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	dbURL := "postgres://delivery-service_user:delivery-service_password@localhost:5433/delivery-service_delivery-service_test?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	_, err = pool.Exec(ctx, "TRUNCATE TABLE restaurants RESTART IDENTITY CASCADE;")
	if err != nil {
		t.Fatalf("failed to truncate restaurants table: %v", err)
	}

	teardown := func() {
		_, _ = pool.Exec(ctx, "TRUNCATE TABLE restaurants RESTART IDENTITY CASCADE;")
		pool.Close()
	}

	return pool, teardown
}

func TestRepository(t *testing.T) {
	pool, teardown := setupTestDB(t)
	defer teardown()

	repo := NewRepository(pool)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("success create and get by id", func(t *testing.T) {
		rest := &domain.Restaurant{
			Name:       "Додо Пицца",
			Address:    "Невский пр., 20",
			IsActive:   true,
			WebhookURL: "http://localhost:8081/webhook",
		}

		err := repo.Create(ctx, rest)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if rest.ID == 0 {
			t.Errorf("expected ID to be set")
		}

		fetched, err := repo.GetByID(ctx, rest.ID)
		if err != nil {
			t.Fatalf("expected to get restaurant, got: %v", err)
		}
		if fetched.Name != rest.Name {
			t.Errorf("expected name %s, got %s", rest.Name, fetched.Name)
		}
	})

	t.Run("list active filters inactive restaurants", func(t *testing.T) {
		activeRest := &domain.Restaurant{
			Name:       "Active Cafe",
			Address:    "Street 1",
			IsActive:   true,
			WebhookURL: "http://active.com",
		}
		inactiveRest := &domain.Restaurant{
			Name:       "Closed Cafe",
			Address:    "Street 2",
			IsActive:   false,
			WebhookURL: "http://inactive.com",
		}

		_ = repo.Create(ctx, activeRest)
		_ = repo.Create(ctx, inactiveRest)

		list, err := repo.ListActive(ctx)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		for _, item := range list {
			if !item.IsActive {
				t.Errorf("expected only active restaurants, got inactive with id %d", item.ID)
			}
		}
	})

	t.Run("get by id not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 999999)
		if !errors.Is(err, domain.ErrRestaurantNotFound) {
			t.Errorf("expected ErrRestaurantNotFound, got: %v", err)
		}
	})
}
