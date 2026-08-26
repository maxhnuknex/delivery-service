package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"delivery-service/internal/domain"
)

type RestaurantRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Restaurant, error)
}

type OrderStatusUpdater interface {
	UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus) error
}

type NotifierWorker struct {
	orderQueue     <-chan *domain.Order
	restaurantRepo RestaurantRepository
	statusUpdater  OrderStatusUpdater
	httpClient     *http.Client
}

func NewNotifierWorker(
	orderQueue <-chan *domain.Order,
	restaurantRepo RestaurantRepository,
	statusUpdater OrderStatusUpdater,
) *NotifierWorker {
	return &NotifierWorker{
		orderQueue:     orderQueue,
		restaurantRepo: restaurantRepo,
		statusUpdater:  statusUpdater,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Start запускает воркер в бесконечном цикле до отмены контекста
func (w *NotifierWorker) Start(ctx context.Context) {
	log.Println("[Worker] Notification worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("[Worker] Notification worker stopped gracefully")
			return
		case order, ok := <-w.orderQueue:
			if !ok {
				log.Println("[Worker] Order queue channel closed")
				return
			}
			w.processOrder(ctx, order)
		}
	}
}

func (w *NotifierWorker) processOrder(ctx context.Context, order *domain.Order) {
	rest, err := w.restaurantRepo.GetByID(ctx, order.RestaurantID)
	if err != nil {
		log.Printf("[Worker] Failed to fetch restaurant %d for order %d: %v", order.RestaurantID, order.ID, err)
		return
	}

	if rest.WebhookURL == "" {
		log.Printf("[Worker] Restaurant %d has no webhook URL, skipping", rest.ID)
		return
	}

	payload, err := json.Marshal(order)
	if err != nil {
		log.Printf("[Worker] Failed to marshal order %d: %v", order.ID, err)
		return
	}

	// Делаем до 3 попыток отправки с задержкой
	var sendErr error
	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, rest.WebhookURL, bytes.NewBuffer(payload))
		if err != nil {
			sendErr = err
			break
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := w.httpClient.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				log.Printf("[Worker] Successfully notified restaurant %d about order %d", rest.ID, order.ID)
				_ = w.statusUpdater.UpdateStatus(ctx, order.ID, domain.OrderStatusAccepted)
				return
			}
			sendErr = fmt.Errorf("bad status code: %d", resp.StatusCode)
		} else {
			sendErr = err
		}

		log.Printf("[Worker] Attempt %d failed to send order %d to %s: %v", attempt, order.ID, rest.WebhookURL, sendErr)
		time.Sleep(1 * time.Second)
	}

	log.Printf("[Worker] Giving up on order %d notification after 3 attempts", order.ID)
}
