package main

import (
	"net/http"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/config"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	mux.Post("/search-availability-json", handlers.Repo.SearchAvailabilityJson)
	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/book-room", handlers.Repo.BookRoom)

	mux.Get("/make-reservation", handlers.Repo.MakeReservation)
	mux.Post("/make-reservation", handlers.Repo.PostMakeReservation)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)

	mux.Get("/auth/login", handlers.Repo.ShowLogin)
	mux.Post("/auth/login", handlers.Repo.PostShowLogin)
	mux.Get("/auth/logout", handlers.Repo.Logout)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(Auth)

		mux.Get("/dashboard", handlers.Repo.AdminDashboard)
		mux.Get("/new-reservations", handlers.Repo.AdminNewReservations)
		mux.Get("/all-reservations", handlers.Repo.AdminAllReservations)
		mux.Get("/reservations-calendar", handlers.Repo.AdminReservationsCalendar)
		mux.Post("/reservations-calendar", handlers.Repo.AdminPostReservationsCalendar)
		mux.Get("/reservations/{src}/{id}", handlers.Repo.AdminShowReservation)
		mux.Post("/reservations/{src}/{id}", handlers.Repo.AdminPostShowReservation)
		mux.Get("/process-reservation/{src}/{id}", handlers.Repo.AdminProcessReservation)
		mux.Get("/delete-reservation/{src}/{id}", handlers.Repo.AdminDeleteReservation)
	})

	// Serve static files
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}