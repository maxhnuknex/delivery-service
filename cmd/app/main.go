package main

import (
	"log"

	"delivery-service/internal/repository/postgres"
	menuRepo "delivery-service/internal/repository/postgres/menu"
	orderRepo "delivery-service/internal/repository/postgres/order"
	restaurantRepo "delivery-service/internal/repository/postgres/restaurant"
	userRepo "delivery-service/internal/repository/postgres/user"
)

func main() {
	dbURL := "postgres://delivery-service_user:delivery-service_password@localhost:5432/delivery-service_delivery-service?sslmode=disable"

	log.Println("Connecting to PostgreSQL...")

	storage, err := postgres.NewStorage(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	log.Println("Successfully connected to PostgreSQL! The database is ready.")

	userRepo := userRepo.NewRepository(storage.Pool)
	menuRepo := menuRepo.NewRepository(storage.Pool)
	orderRepo := orderRepo.NewRepository(storage.Pool)
	restaurantRepo := restaurantRepo.NewRepository(storage.Pool)

}
