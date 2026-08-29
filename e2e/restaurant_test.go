package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"delivery-service/internal/domain"
)

func TestE2E_RestaurantFlow(t *testing.T) {
	apiURL := getAPIURL()

	var createdRestID int64

	t.Run("1. create restaurant", func(t *testing.T) {
		reqBody := domain.CreateRestaurantDTO{
			Name:       "Додо Пицца E2E",
			Type:       domain.EstablishmentTypeRestaurant,
			Address:    "Невский пр., 20",
			WebhookURL: "http://delivery-service-mock_restaurant_e2e-1:8081/webhook",
		}

		bodyBytes, _ := json.Marshal(reqBody)
		resp, err := client.Post(apiURL+"/api/v1/restaurants", "application/json", bytes.NewBuffer(bodyBytes))
		if err != nil {
			t.Fatalf("failed to execute request: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Fatalf("failed to close body: %v", err)
			}
		}()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200/201, got %d", resp.StatusCode)
		}

		var rest domain.Restaurant
		if err := json.NewDecoder(resp.Body).Decode(&rest); err != nil {
			t.Fatalf("failed to decode created restaurant: %v", err)
		}
		if rest.ID == 0 {
			t.Fatalf("expected valid restaurant ID, got 0")
		}
		createdRestID = rest.ID
	})

	t.Run("2. add menu item", func(t *testing.T) {
		itemDTO := domain.CreateMenuItemDTO{
			Name:          "Пепперони",
			Description:   "Острая пицца с салями",
			Price:         550,
			StockQuantity: 10,
		}

		bodyBytes, _ := json.Marshal(itemDTO)
		url := fmt.Sprintf("%s/api/v1/restaurants/%d/menu", apiURL, createdRestID)
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
		if err != nil {
			t.Fatalf("failed to execute request: %v", err)
		}

		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Fatalf("failed to close body: %v", err)
			}
		}()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200/201, got %d", resp.StatusCode)
		}
	})

	t.Run("3. get restaurant menu", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/restaurants/%d/menu", apiURL, createdRestID)
		resp, err := client.Get(url)
		if err != nil {
			t.Fatalf("failed to get menu: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Fatalf("failed to close body: %v", err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var menu []*domain.MenuItem
		if err := json.NewDecoder(resp.Body).Decode(&menu); err != nil {
			t.Fatalf("failed to decode menu: %v", err)
		}

		if len(menu) == 0 {
			t.Fatalf("expected menu items, got empty list")
		}
	})
}
