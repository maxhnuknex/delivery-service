package user

import (
	"context"
	"errors"
	"fmt"

	"delivery-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (name, phone, address)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query, user.Name, user.Phone, user.Address).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return fmt.Errorf("user repository - create: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `SELECT id, name, phone, address, created_at FROM users WHERE id = $1`

	u := &domain.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&u.ID, &u.Name, &u.Phone, &u.Address, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("user repository - get by id: %w", err)
	}
	return u, nil
}

func (r *Repository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `SELECT id, name, phone, address, created_at FROM users WHERE phone = $1`

	u := &domain.User{}
	err := r.pool.QueryRow(ctx, query, phone).Scan(&u.ID, &u.Name, &u.Phone, &u.Address, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("user repository - get by phone: %w", err)
	}
	return u, nil
}
