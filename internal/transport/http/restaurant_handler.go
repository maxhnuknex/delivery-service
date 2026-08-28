package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"delivery-service/internal/domain"

	"github.com/go-chi/chi/v5"
)

type RestaurantService interface {
	ListActiveRestaurants(ctx context.Context, estType domain.EstablishmentType) ([]*domain.Restaurant, error)
	GetMenu(ctx context.Context, restaurantID int64) ([]*domain.MenuItem, error)
	CreateRestaurant(ctx context.Context, dto domain.CreateRestaurantDTO) (*domain.Restaurant, error)
	CreateMenu(ctx context.Context, restID int64, dto domain.CreateMenuItemDTO) (*domain.MenuItem, error)
}

type RestaurantHandler struct {
	service RestaurantService
}

func NewRestaurantHandler(service RestaurantService) *RestaurantHandler {
	return &RestaurantHandler{service: service}
}

// GET /api/v1/restaurants?type=SHOP
func (h *RestaurantHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	typeParam := strings.ToUpper(r.URL.Query().Get("type"))
	estType := domain.EstablishmentType(typeParam)

	list, err := h.service.ListActiveRestaurants(r.Context(), estType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if list == nil {
		list = make([]*domain.Restaurant, 0)
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

	if menu == nil {
		menu = make([]*domain.MenuItem, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(menu)
}

// POST /api/v1/restaurants
func (h *RestaurantHandler) CreateRestaurant(w http.ResponseWriter, r *http.Request) {
	var dto domain.CreateRestaurantDTO

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	rest, err := h.service.CreateRestaurant(r.Context(), dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(rest)
}

// POST /api/v1/restaurants/{id}/menu
func (h *RestaurantHandler) CreateMenu(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	restID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return
	}

	var dto domain.CreateMenuItemDTO

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	menu, err := h.service.CreateMenu(r.Context(), restID, dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(menu)
}
