package user

import (
	"context"
	"strings"
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

	// Очищаем таблицу перед запуском теста
	_, err = pool.Exec(ctx, "TRUNCATE TABLE users RESTART IDENTITY CASCADE;")
	if err != nil {
		t.Fatalf("failed to truncate users table: %v", err)
	}

	teardown := func() {
		_, _ = pool.Exec(ctx, "TRUNCATE TABLE users RESTART IDENTITY CASCADE;")
		pool.Close()
	}

	return pool, teardown
}

func TestRepository_Create(t *testing.T) {
	pool, teardown := setupTestDB(t)
	defer teardown()

	repo := NewRepository(pool)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("success create", func(t *testing.T) {
		user := &domain.User{
			Name:    "Ivan Pashkin",
			Phone:   "+79001234567",
			Address: "SPb, Nevsky 1",
		}

		err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if user.ID == 0 {
			t.Errorf("expected ID to be assigned, got 0")
		}
	})

	t.Run("duplicate phone error", func(t *testing.T) {
		user1 := &domain.User{Name: "User 1", Phone: "+79990000000", Address: "Address 1"}
		user2 := &domain.User{Name: "User 2", Phone: "+79990000000", Address: "Address 2"}

		_ = repo.Create(ctx, user1)

		err := repo.Create(ctx, user2)
		if err == nil {
			t.Fatalf("expected error due to unique constraint, got nil")
		}
		if !strings.Contains(err.Error(), "SQLSTATE 23505") {
			t.Errorf("expected unique constraint violation error, got: %v", err)
		}
	})
}
