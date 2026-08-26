package domain

import (
	"errors"
	"time"
)

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrInvalidStatus = errors.New("invalid order status transition")
)

type OrderStatus string

const (
	OrderStatusCreated          OrderStatus = "CREATED"
	OrderStatusAccepted         OrderStatus = "ACCEPTED"
	OrderStatusInProgress       OrderStatus = "IN_PROGRESS"
	OrderStatusReadyForDelivery OrderStatus = "READY_FOR_DELIVERY"
	OrderStatusDelivered        OrderStatus = "DELIVERED"
	OrderStatusCancelled        OrderStatus = "CANCELLED"
)

type OrderItem struct {
	ID         int64 `json:"id"`
	OrderID    int64 `json:"order_id"`
	MenuItemID int64 `json:"menu_item_id"`
	Quantity   int   `json:"quantity"`
	Price      int64 `json:"price"` // Цена в копейках
}

type Order struct {
	ID              int64       `json:"id"`
	UserID          int64       `json:"user_id"`
	RestaurantID    int64       `json:"restaurant_id"`
	Status          OrderStatus `json:"status"`
	DeliveryAddress string      `json:"delivery_address"`
	TotalPrice      int64       `json:"total_price"`
	Items           []OrderItem `json:"items"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type CreateOrderDTO struct {
	UserID          int64
	RestaurantID    int64
	DeliveryAddress string
	Items           []CreateOrderItemDTO
}

type CreateOrderItemDTO struct {
	MenuItemID int64
	Quantity   int
}
