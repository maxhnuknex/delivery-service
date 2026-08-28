package domain

import (
	"errors"
	"time"
)

type EstablishmentType string

const (
	EstablishmentTypeRestaurant EstablishmentType = "RESTAURANT"
	EstablishmentTypeShop       EstablishmentType = "SHOP"
	EstablishmentTypeCafe       EstablishmentType = "CAFE"
)

var (
	ErrRestaurantNotFound = errors.New("restaurant not found")
	ErrRestaurantInactive = errors.New("restaurant is inactive")
)

type Restaurant struct {
	ID         int64             `json:"id"`
	Name       string            `json:"name"`
	Type       EstablishmentType `json:"type"`
	Address    string            `json:"address"`
	IsActive   bool              `json:"is_active"`
	WebhookURL string            `json:"webhook_url"`
	CreatedAt  time.Time         `json:"created_at"`
}
