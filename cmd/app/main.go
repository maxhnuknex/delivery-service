package main

import (
	"log"

	"delivery-service/internal/repository/postgres"
)

func main() {
	// DSN для локального запуска (без Docker для приложения, БД крутится в Docker на порту 5432)
	dbURL := "postgres://delivery-service_user:delivery-service_password@localhost:5432/delivery-service_delivery-service?sslmode=disable"

	log.Println("Connecting to PostgreSQL...")

	storage, err := postgres.NewStorage(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	log.Println("Successfully connected to PostgreSQL! The database is ready.")
}
