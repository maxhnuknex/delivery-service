package domain

// Для пользователя
type CreateUserDTO struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

// Для ресторана
type CreateRestaurantDTO struct {
	Name       string            `json:"name"`
	Type       EstablishmentType `json:"type"`
	Address    string            `json:"address"`
	WebhookURL string            `json:"webhook_url"`
}

// Для меню
type CreateMenuItemDTO struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Price         int64  `json:"price"`
	StockQuantity int    `json:"stock_quantity"`
}

type CreateOrderDTO struct {
	UserID          int64                `json:"user_id"`
	RestaurantID    int64                `json:"restaurant_id"`
	DeliveryAddress string               `json:"delivery_address"`
	Items           []CreateOrderItemDTO `json:"items"`
}

type CreateOrderItemDTO struct {
	MenuItemID int64 `json:"menu_item_id"`
	Quantity   int   `json:"quantity"`
}

type UpdateOrderStatusDTO struct {
	Status OrderStatus `json:"status"`
}
