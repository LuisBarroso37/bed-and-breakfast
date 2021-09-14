package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/config"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/forms"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/helpers"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/render"
)

type Repository struct {
	App *config.AppConfig
}

// Reposiotry used by the handlers
var Repo *Repository

// Create a new repository
func NewRepository(config *config.AppConfig) *Repository {
	return &Repository{ App: config }
}

// Sets the repository for the handlers
func SetRepository(repo *Repository) {
	Repo = repo
}

// Home is the home page handler
func (repo *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the about page handler
func (repo *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "about.page.tmpl", &models.TemplateData{})
}

// MakeReservation is the make a reservation page handler
func (repo *Repository) MakeReservation(w http.ResponseWriter, r *http.Request) {
	// Send an empty reservation model as Data in TemplateData
	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// Handles the creation of a reservation
func (repo *Repository) CreateReservation(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Save form data in a Reservation struct
	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName: r.Form.Get("last_name"),
		Email: r.Form.Get("email"),
		Phone: r.Form.Get("phone"),
	}

	// Validate form data and add any errors that might exist to `form` variable
	form := forms.New(r.PostForm)
	form.RequiredFields("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 2)
	form.IsEmail("email")

	// Rerender make reservation form with updated error information
	if !form.IsValid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})

		return
	}

	// Store form data in the app configuration
	repo.App.Session.Put(r.Context(), "reservation", reservation)

	// Redirect user to reservation summary page
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// SearchAvailability is the search availability page handler
func (repo *Repository) SearchAvailability(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

// Accepts form data for search availability and returns a response
func (repo *Repository) PostSearchAvailability(w http.ResponseWriter, r *http.Request) {
	// Retrieve values from submitted form
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	
	w.Write([]byte(fmt.Sprintf("Start date is %s and end date is %s", start, end)))
}

type jsonResponse struct {
	OK 			bool 		`json:"ok"`
	Message string 	`json:"message"`
}

// Accepts form data for search availability and returns a JSON response
func (repo *Repository) AvailabilityJson(w http.ResponseWriter, r *http.Request) {
	// Response to send back as JSON
	res := jsonResponse{
		OK: true,
		Message: "Available!",
	}

	// Convert response to JSON
	json, err := json.Marshal(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Send back the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

// Generals is the Generals quarters room page handler
func (repo *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "generals-quarters.page.tmpl", &models.TemplateData{})
}

// Majors is the Majors quarters room page handler
func (repo *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "majors-suite.page.tmpl", &models.TemplateData{})
}

// Contact is the contact page handler
func (repo *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// ReservationSummary is the reservation summary page handler
func (repo *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	// Get make reservation form data from Session object
	reservation, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation) // Type assertion
	if !ok {
		repo.App.ErrorLog.Println("Cannot get item from session")

		// Show error message to user and redirect them to home page
		repo.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Delete reservation data from session object
	repo.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.RenderTemplate(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
}