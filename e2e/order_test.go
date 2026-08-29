package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"delivery-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

func createTestUserFixture(t *testing.T) int64 {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://test_user:test_password@localhost:5433/delivery-service_delivery-service_test?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to connect to test db for fixtures: %v", err)
	}
	defer pool.Close()

	var userID int64
	phone := fmt.Sprintf("+7999%d", time.Now().UnixNano()%10000000)
	err = pool.QueryRow(ctx, `
		INSERT INTO users (name, phone, address)
		VALUES ('E2E Покупатель', $1, 'ул. Пушкина, 10')
		RETURNING id
	`, phone).Scan(&userID)
	if err != nil {
		t.Fatalf("failed to insert test user fixture: %v", err)
	}

	return userID
}

func TestE2E_OrderLifecycle(t *testing.T) {
	apiURL := getAPIURL()

	// 1. Создаем пользователя напрямую в БД E2E
	userID := createTestUserFixture(t)

	// 2. Создаем ресторан через публичный API
	restDTO := domain.CreateRestaurantDTO{
		Name:       "Бургерная E2E",
		Type:       domain.EstablishmentTypeRestaurant,
		Address:    "ул. Садовая, 10",
		WebhookURL: "http://delivery-service-mock_restaurant_e2e-1:8081/webhook",
	}
	restBody, _ := json.Marshal(restDTO)
	respRest, err := client.Post(apiURL+"/api/v1/restaurants", "application/json", bytes.NewBuffer(restBody))
	if err != nil {
		t.Fatalf("failed to create restaurant: %v", err)
	}

	defer func() {
		if err := respRest.Body.Close(); err != nil {
			t.Fatalf("failed to close body: %v", err)
		}
	}()

	if respRest.StatusCode != http.StatusCreated && respRest.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(respRest.Body)
		t.Fatalf("failed to create restaurant, code: %d, body: %s", respRest.StatusCode, string(b))
	}

	var rest domain.Restaurant
	_ = json.NewDecoder(respRest.Body).Decode(&rest)

	// 3. Добавляем позицию в меню
	menuDTO := domain.CreateMenuItemDTO{
		Name:          "Чизбургер E2E",
		Price:         350,
		StockQuantity: 10,
	}
	menuBody, _ := json.Marshal(menuDTO)
	respMenu, err := client.Post(fmt.Sprintf("%s/api/v1/restaurants/%d/menu", apiURL, rest.ID), "application/json", bytes.NewBuffer(menuBody))
	if err != nil {
		t.Fatalf("failed to add menu item: %v", err)
	}
	defer func() {
		if err := respMenu.Body.Close(); err != nil {
			t.Fatalf("failed to close body: %v", err)
		}
	}()

	var item domain.MenuItem
	_ = json.NewDecoder(respMenu.Body).Decode(&item)

	// 4. Оформляем заказ
	orderDTO := domain.CreateOrderDTO{
		UserID:          userID,
		RestaurantID:    rest.ID,
		DeliveryAddress: "ул. Пушкина, 10",
		Items: []domain.CreateOrderItemDTO{
			{MenuItemID: item.ID, Quantity: 2},
		},
	}
	orderBody, _ := json.Marshal(orderDTO)
	respOrder, err := client.Post(apiURL+"/api/v1/orders", "application/json", bytes.NewBuffer(orderBody))
	if err != nil {
		t.Fatalf("failed to execute order request: %v", err)
	}
	defer func() {
		if err := respOrder.Body.Close(); err != nil {
			t.Fatalf("failed to close body: %v", err)
		}
	}()

	if respOrder.StatusCode != http.StatusCreated && respOrder.StatusCode != http.StatusOK {
		rawResp, _ := io.ReadAll(respOrder.Body)
		t.Fatalf("expected order created status, got %d. Response: %s", respOrder.StatusCode, string(rawResp))
	}

	var createdOrder domain.Order
	if err := json.NewDecoder(respOrder.Body).Decode(&createdOrder); err != nil {
		t.Fatalf("failed to decode created order: %v", err)
	}

	if createdOrder.TotalPrice != 700 {
		t.Errorf("expected total price 700, got %d", createdOrder.TotalPrice)
	}
	if createdOrder.Status != domain.OrderStatusCreated {
		t.Errorf("expected status '%s', got '%s'", domain.OrderStatusCreated, createdOrder.Status)
	}
}
