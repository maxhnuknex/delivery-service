package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(restHandler *RestaurantHandler, orderHandler *OrderHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		// Каталог и меню
		r.Route("/restaurants", func(r chi.Router) {
			r.Get("/", restHandler.ListActive)
			r.Get("/{id}/menu", restHandler.GetMenu)
		})

		// Заказы
		r.Route("/orders", func(r chi.Router) {
			r.Post("/", orderHandler.CreateOrder)
			r.Get("/{id}", orderHandler.GetOrderByID)
			r.Patch("/{id}/status", orderHandler.UpdateStatus)
		})

		// История пользователя
		r.Route("/users", func(r chi.Router) {
			r.Get("/{id}/orders", orderHandler.ListUserOrders)
		})
	})

	return r
}
