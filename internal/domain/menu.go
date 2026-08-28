package domain

import (
	"errors"
	"time"
)

var (
	ErrMenuItemNotFound  = errors.New("menu item not found")
	ErrInsufficientStock = errors.New("insufficient stock quantity")
)

type MenuItem struct {
	ID            int64     `json:"id"`
	RestaurantID  int64     `json:"restaurant_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         int64     `json:"price"`
	StockQuantity int       `json:"stock_quantity"`
	IsAvailable   bool      `json:"is_available"`
	CreatedAt     time.Time `json:"created_at"`
}
