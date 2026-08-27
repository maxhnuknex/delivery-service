package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"delivery-service/internal/domain"

	"github.com/go-chi/chi/v5"
)

type OrderService interface {
	CreateOrder(ctx context.Context, dto domain.CreateOrderDTO) (*domain.Order, error)
	GetByID(ctx context.Context, id int64) (*domain.Order, error)
	ListUserOrders(ctx context.Context, userID int64) ([]*domain.Order, error)
	UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus) error
}

type OrderHandler struct {
	service OrderService
}

func NewOrderHandler(service OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// POST /api/v1/orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var dto domain.CreateOrderDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	createdOrder, err := h.service.CreateOrder(r.Context(), dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(createdOrder)
}

// GET /api/v1/orders/{id}
func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	resOrder, err := h.service.GetByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resOrder)
}

// GET /api/v1/users/{id}/orders
func (h *OrderHandler) ListUserOrders(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	orders, err := h.service.ListUserOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if orders == nil {
		orders = make([]*domain.Order, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(orders)
}

// PATCH /api/v1/orders/{id}/status
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	var req domain.UpdateOrderStatusDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateStatus(r.Context(), orderID, req.Status); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
