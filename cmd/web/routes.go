package main

import (
	"net/http"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/config"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/handlers"
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
	mux.Get("/generals-quarters", handlers.Repo.Generals)
	mux.Get("/majors-suite", handlers.Repo.Majors)
	mux.Get("/contact", handlers.Repo.Contact)

	mux.Get("/search-availability", handlers.Repo.SearchAvailability)
	mux.Post("/search-availability", handlers.Repo.PostSearchAvailability)
	mux.Post("/search-availability-json", handlers.Repo.AvailabilityJson)

	mux.Get("/make-reservation", handlers.Repo.MakeReservation)
	mux.Post("/make-reservation", handlers.Repo.CreateReservation)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)

	// Serve static files
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}