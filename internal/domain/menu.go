package domain

import (
	"errors"
	"time"
)

var ErrMenuItemNotFound = errors.New("menu item not found")

type MenuItem struct {
	ID           int64
	RestaurantID int64
	Name         string
	Description  string
	Price        int64
	IsAvailable  bool
	CreatedAt    time.Time
}
