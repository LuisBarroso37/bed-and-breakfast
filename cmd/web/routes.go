package main

import (
	"net/http"

	"github.com/LuisBarroso37/bed-and-breakfast/pkg/config"
	"github.com/LuisBarroso37/bed-and-breakfast/pkg/handlers"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Routes for our server
func routes(app *config.AppConfig) http.Handler {
	// Create router
	mux := chi.NewRouter()

	// Set middlewares
	mux.Use(middleware.Recoverer)
	mux.Use(CreateCsrfHandler)
	mux.Use(SessionLoad)

	// Routes
	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)

	return mux
}