package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"delivery-service/internal/domain"
	menuRepo "delivery-service/internal/repository/postgres/menu"
	orderRepo "delivery-service/internal/repository/postgres/order"
	restRepo "delivery-service/internal/repository/postgres/restaurant"
	userRepo "delivery-service/internal/repository/postgres/user"
	orderSvc "delivery-service/internal/service/order"
	restSvc "delivery-service/internal/service/restaurant"
	transportHTTP "delivery-service/internal/transport/http"
	"delivery-service/internal/worker"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://delivery-service_user:delivery-service_password@localhost:5432/delivery-service_delivery-service?sslmode=disable"
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	uRepo := userRepo.NewRepository(pool)
	rRepo := restRepo.NewRepository(pool)
	mRepo := menuRepo.NewRepository(pool)
	oRepo := orderRepo.NewRepository(pool)

	orderEvents := make(chan *domain.Order, 100)

	restaurantService := restSvc.NewService(rRepo, mRepo)
	orderService := orderSvc.NewService(uRepo, rRepo, mRepo, oRepo, orderEvents)

	notifier := worker.NewNotifierWorker(orderEvents, rRepo, oRepo)
	go notifier.Start(ctx)

	restHandler := transportHTTP.NewRestaurantHandler(restaurantService)
	orderHandler := transportHTTP.NewOrderHandler(orderService)
	router := transportHTTP.NewRouter(restHandler, orderHandler)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("Server running on port :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	log.Println("Server exited successfully")
}
