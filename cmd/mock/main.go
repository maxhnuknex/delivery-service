package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func apiBaseURL() string {
	value := os.Getenv("API_BASE_URL")
	if value == "" {
		log.Fatal("API_BASE_URL is not set")
	}
	return value
}

func webhookURL() string {
	if value := os.Getenv("MOCK_WEBHOOK_URL"); value != "" {
		return value
	}
	return "http://mock-restaurant:8081/webhook"
}

type menuItem struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Price         int64  `json:"price"`
	StockQuantity int    `json:"stock_quantity"`
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func statusDelay() time.Duration {
	seconds, err := strconv.Atoi(env("STATUS_DELAY_SECONDS", "15"))
	if err != nil || seconds < 1 {
		seconds = 15
	}
	return time.Duration(seconds) * time.Second
}

func main() {
	// 1. Ждем, пока основное API поднимется
	time.Sleep(3 * time.Second)

	// 2. Регистрируем себя в сервисе
	restID := registerRestaurant()
	if restID > 0 {
		createMenuItems(restID)
	}

	// 3. Запускаем прослушивание вебхуков
	http.HandleFunc("/webhook", handleWebhook)

	mockPort := os.Getenv("MOCK_PORT")
	if mockPort == "" {
		log.Fatal("MOCK_PORT is not set")
	}
	log.Printf("Mock Restaurant запущен на порту :%s...", mockPort)
	log.Fatal(http.ListenAndServe(":"+mockPort, nil))
}

func registerRestaurant() int64 {
	payload := map[string]interface{}{
		"name":        env("MOCK_NAME", "Mock Restaurant"),
		"type":        env("MOCK_TYPE", "RESTAURANT"),
		"address":     env("MOCK_ADDRESS", "ул. Интеграционная, 1"),
		"webhook_url": webhookURL(),
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(apiBaseURL()+"/restaurants", "application/json", bytes.NewBuffer(body))
	if err != nil || resp.StatusCode != http.StatusCreated {
		log.Printf("[MOCK] Ошибка регистрации ресторана (возможно, уже существует): %v", err)
		return 0
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("MOCK no close() body")
		}
	}()
	var result struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[MOCK] invalid json")
		return 0
	}
	log.Printf("[MOCK] Ресторан успешно зарегистрирован с ID: %d", result.ID)
	return result.ID
}

func createMenuItems(restID int64) {
	items := []menuItem{}
	if err := json.Unmarshal([]byte(env("MOCK_MENU", "[]")), &items); err != nil {
		log.Printf("[MOCK] Не удалось прочитать MOCK_MENU: %v", err)
		return
	}

	for _, item := range items {
		body, err := json.Marshal(item)
		if err != nil {
			continue
		}
		url := fmt.Sprintf("%s/restaurants/%d/menu", apiBaseURL(), restID)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Printf("[MOCK] Не удалось добавить %q: %v", item.Name, err)
			continue
		}
		if err := resp.Body.Close(); err != nil {
			log.Printf("[MOCK] Не удалось закрыть ответ: %v", err)
		}
		if resp.StatusCode == http.StatusCreated {
			log.Printf("[MOCK] Добавлена позиция меню: %s", item.Name)
		}
	}
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		OrderID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	log.Printf("[MOCK] Получен заказ #%d. Начинаем готовить...\n", payload.OrderID)
	w.WriteHeader(http.StatusOK)

	go processOrder(payload.OrderID)
}

func processOrder(orderID int64) {
	statuses := []string{"IN_PROGRESS", "READY_FOR_DELIVERY", "DELIVERED"}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, status := range statuses {
		time.Sleep(statusDelay())

		body, _ := json.Marshal(map[string]string{"status": status})
		url := fmt.Sprintf("%s/orders/%d/status", apiBaseURL(), orderID)

		req, _ := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[MOCK] Ошибка сети при смене статуса %s: %v\n", status, err)
			continue
		}

		if err := resp.Body.Close(); err != nil {
			log.Printf("MOCK no close() body")
		}

		if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
			log.Printf("[MOCK] Статус заказа #%d изменен на %s\n", orderID, status)
		} else {
			log.Printf("[MOCK] Ошибка API! API вернул код %d при попытке поставить статус %s\n", resp.StatusCode, status)
		}
	}
}
