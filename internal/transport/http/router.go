package http

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed web/*
var webFiles embed.FS

func NewRouter(restHandler *RestaurantHandler, orderHandler *OrderHandler, userHandler *UserHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		// Каталог и меню
		r.Route("/restaurants", func(r chi.Router) {
			r.Get("/", restHandler.ListActive)
			r.Get("/{id}/menu", restHandler.GetMenu)
			r.Post("/", restHandler.CreateRestaurant)
			r.Post("/{id}/menu", restHandler.CreateMenu)
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

		r.Post("/users", userHandler.CreateUser)
	})

	webRoot, err := fs.Sub(webFiles, "web")
	if err != nil {
		panic("embedded web files are unavailable: " + err.Error())
	}
	r.Handle("/*", http.FileServer(http.FS(webRoot)))

	return r
}
