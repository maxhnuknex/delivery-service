package http

import (
	"context"
	"encoding/json"
	"net/http"

	"delivery-service/internal/domain"
)

type UserService interface {
	Create(ctx context.Context, dto domain.CreateUserDTO) (*domain.User, error)
}

type UserHandler struct {
	service UserService
}

func New(service UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// POST /api/v1/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var dto domain.CreateUserDTO

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	user, err := h.service.Create(r.Context(), dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}
