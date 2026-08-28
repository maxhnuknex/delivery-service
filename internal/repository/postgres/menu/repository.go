package menu

import (
	"context"
	"fmt"

	"delivery-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, item *domain.MenuItem) error {
	query := `
		INSERT INTO menu_items (restaurant_id, name, description, price, stock_quantity, is_available)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(
		ctx, query,
		item.RestaurantID, item.Name, item.Description, item.Price, item.StockQuantity, item.IsAvailable,
	).Scan(&item.ID, &item.CreatedAt)
	if err != nil {
		return fmt.Errorf("menu repository - create: %w", err)
	}
	return nil
}

func (r *Repository) GetMenu(ctx context.Context, restaurantID int64) ([]*domain.MenuItem, error) {
	query := `
		SELECT id, restaurant_id, name, description, price, stock_quantity, is_available, created_at
		FROM menu_items
		WHERE restaurant_id = $1 AND is_available = true
		ORDER BY id ASC
	`
	rows, err := r.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("menu repository - get menu: %w", err)
	}
	defer rows.Close()

	var items []*domain.MenuItem
	for rows.Next() {
		item := &domain.MenuItem{}
		err := rows.Scan(
			&item.ID, &item.RestaurantID, &item.Name, &item.Description,
			&item.Price, &item.StockQuantity, &item.IsAvailable, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("menu repository - scan menu item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) GetMenuItemsByIDs(ctx context.Context, ids []int64) ([]*domain.MenuItem, error) {
	query := `
		SELECT id, restaurant_id, name, description, price, stock_quantity, is_available, created_at
		FROM menu_items
		WHERE id = ANY($1)
	`
	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("menu repository - get by ids: %w", err)
	}
	defer rows.Close()

	var items []*domain.MenuItem
	for rows.Next() {
		item := &domain.MenuItem{}
		err := rows.Scan(
			&item.ID, &item.RestaurantID, &item.Name, &item.Description,
			&item.Price, &item.StockQuantity, &item.IsAvailable, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("menu repository - scan item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}
