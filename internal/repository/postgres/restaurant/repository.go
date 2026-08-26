package restaurant

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

func (r *Repository) Create(ctx context.Context, rest *domain.Restaurant) error {
	query := `
		INSERT INTO restaurants (name, address, is_active, webhook_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query, rest.Name, rest.Address, rest.IsActive, rest.WebhookURL).
		Scan(&rest.ID, &rest.CreatedAt)
	if err != nil {
		return fmt.Errorf("restaurant repository - create: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*domain.Restaurant, error) {
	query := `
		SELECT id, name, address, is_active, webhook_url, created_at
		FROM restaurants
		WHERE id = $1
	`
	rest := &domain.Restaurant{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rest.ID, &rest.Name, &rest.Address, &rest.IsActive, &rest.WebhookURL, &rest.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRestaurantNotFound
		}
		return nil, fmt.Errorf("restaurant repository - get by id: %w", err)
	}
	return rest, nil
}

func (r *Repository) ListActive(ctx context.Context) ([]*domain.Restaurant, error) {
	query := `
		SELECT id, name, address, is_active, webhook_url, created_at
		FROM restaurants
		WHERE is_active = true
		ORDER BY id ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("restaurant repository - list active: %w", err)
	}
	defer rows.Close()

	var list []*domain.Restaurant
	for rows.Next() {
		rest := &domain.Restaurant{}
		err := rows.Scan(
			&rest.ID, &rest.Name, &rest.Address, &rest.IsActive, &rest.WebhookURL, &rest.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("restaurant repository - scan: %w", err)
		}
		list = append(list, rest)
	}
	return list, nil
}
