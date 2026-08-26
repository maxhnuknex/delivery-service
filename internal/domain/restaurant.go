package domain

import (
	"errors"
	"time"
)

var (
	ErrRestaurantNotFound = errors.New("restaurant not found")
)

type Restaurant struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	IsActive   bool      `json:"is_active"`
	WebhookURL string    `json:"webhook_url"`
	CreatedAt  time.Time `json:"created_at"`
}
