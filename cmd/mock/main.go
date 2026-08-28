package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const apiBaseURL = "http://api:8080/api/v1"

func main() {
	// 1. Ждем, пока основное API поднимется
	time.Sleep(3 * time.Second)

	// 2. Регистрируем себя в Delivery Service.Кухне
	restID := registerRestaurant()
	if restID > 0 {
		createMenuItem(restID)
	}

	// 3. Запускаем прослушивание вебхуков
	http.HandleFunc("/webhook", handleWebhook)

	log.Println("Mock Restaurant запущен на порту 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func registerRestaurant() int64 {
	payload := map[string]interface{}{
		"name":        "Mock Restaurant",
		"type":        "RESTAURANT",
		"address":     "ул. Интеграционная, 1",
		"webhook_url": "http://mock-restaurant:8081/webhook",
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(apiBaseURL+"/restaurants", "application/json", bytes.NewBuffer(body))
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

func createMenuItem(restID int64) {
	payload := map[string]interface{}{
		"name":           "Тестовая Пицца",
		"description":    "Пицца для проверки интеграции",
		"price":          500,
		"stock_quantity": 0,
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/restaurants/%d/menu", apiBaseURL, restID)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err == nil && resp.StatusCode == http.StatusCreated {
		log.Println("[MOCK] Тестовое меню успешно добавлено!")
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
	statuses := []string{"COOKING", "DELIVERING", "COMPLETED"}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, status := range statuses {
		time.Sleep(5 * time.Second)

		body, _ := json.Marshal(map[string]string{"status": status})
		url := fmt.Sprintf("%s/orders/%d/status", apiBaseURL, orderID)

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
