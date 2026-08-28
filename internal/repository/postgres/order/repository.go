package order

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

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

func (r *Repository) Create(ctx context.Context, order *domain.Order, deductStock bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("order repo - begin tx: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Printf("order repo - deduct stock: %v", err)
		}
	}()
	// Если это магазин — атомарно списываем остатки
	if deductStock {
		stockQuery := `
			UPDATE menu_items
			SET stock_quantity = stock_quantity - $1
			WHERE id = $2 AND stock_quantity >= $1
		`
		for _, item := range order.Items {
			cmd, err := tx.Exec(ctx, stockQuery, item.Quantity, item.MenuItemID)
			if err != nil {
				return fmt.Errorf("order repo - deduct stock: %w", err)
			}
			// Если условие stock_quantity >= $1 не выполнилось, ни одна строка не обновилась
			if cmd.RowsAffected() == 0 {
				return domain.ErrInsufficientStock
			}
		}
	}

	// Создание заказа
	now := time.Now().UTC()
	orderQuery := `
		INSERT INTO orders (user_id, restaurant_id, status, delivery_address, total_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	err = tx.QueryRow(
		ctx, orderQuery,
		order.UserID, order.RestaurantID, order.Status, order.DeliveryAddress, order.TotalPrice, now, now,
	).Scan(&order.ID)
	if err != nil {
		return fmt.Errorf("order repo - insert order: %w", err)
	}

	// Создание позиций заказа
	itemQuery := `
		INSERT INTO order_items (order_id, menu_item_id, quantity, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	for i := range order.Items {
		order.Items[i].OrderID = order.ID
		err = tx.QueryRow(
			ctx, itemQuery,
			order.Items[i].OrderID, order.Items[i].MenuItemID, order.Items[i].Quantity, order.Items[i].Price,
		).Scan(&order.Items[i].ID)
		if err != nil {
			return fmt.Errorf("order repo - insert item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("order repo - commit tx: %w", err)
	}

	order.CreatedAt = now
	order.UpdatedAt = now
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	orderQuery := `
		SELECT id, user_id, restaurant_id, status, delivery_address, total_price, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	order := &domain.Order{}
	err := r.pool.QueryRow(ctx, orderQuery, id).Scan(
		&order.ID, &order.UserID, &order.RestaurantID, &order.Status,
		&order.DeliveryAddress, &order.TotalPrice, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("order repository - get order: %w", err)
	}

	itemsQuery := `
		SELECT id, order_id, menu_item_id, quantity, price
		FROM order_items
		WHERE order_id = $1
		ORDER BY id ASC
	`
	rows, err := r.pool.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("order repository - get order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.MenuItemID, &item.Quantity, &item.Price); err != nil {
			return nil, fmt.Errorf("order repository - scan order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	res, err := r.pool.Exec(ctx, query, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("order repository - update status: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

func (r *Repository) ListByUserID(ctx context.Context, userID int64) ([]*domain.Order, error) {
	query := `
		SELECT id, user_id, restaurant_id, status, delivery_address, total_price, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("order repository - list by user: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		o := &domain.Order{}
		if err := rows.Scan(&o.ID, &o.UserID, &o.RestaurantID, &o.Status, &o.DeliveryAddress, &o.TotalPrice, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("order repository - scan user order: %w", err)
		}
		orders = append(orders, o)
	}
	return orders, nil
}
