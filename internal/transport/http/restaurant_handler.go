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

type RestaurantService interface {
	ListActiveRestaurants(ctx context.Context) ([]*domain.Restaurant, error)
	GetMenu(ctx context.Context, restaurantID int64) ([]*domain.MenuItem, error)
}

type RestaurantHandler struct {
	service RestaurantService
}

func NewRestaurantHandler(service RestaurantService) *RestaurantHandler {
	return &RestaurantHandler{service: service}
}

// GET /api/v1/restaurants
func (h *RestaurantHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListActiveRestaurants(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

// GET /api/v1/restaurants/{id}/menu
func (h *RestaurantHandler) GetMenu(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	restID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid restaurant id", http.StatusBadRequest)
		return
	}

	menu, err := h.service.GetMenu(r.Context(), restID)
	if err != nil {
		if errors.Is(err, domain.ErrRestaurantNotFound) {
			http.Error(w, "restaurant not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrRestaurantInactive) {
			http.Error(w, "restaurant is inactive", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(menu)
}
