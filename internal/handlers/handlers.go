package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/config"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/driver"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/forms"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/helpers"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/render"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/repository"
	dbrepository "github.com/LuisBarroso37/bed-and-breakfast/internal/repository/db-repository"
)

type Repository struct {
	App *config.AppConfig
	DB repository.DatabaseRepository
}

// Repository used by the handlers
var Repo *Repository

// Create a new repository
func NewRepository(app *config.AppConfig, pool *driver.DB) *Repository {
	return &Repository{ 
		App: app,
		DB: dbrepository.NewPostgresRepository(pool.SQL, app),
	}
}

// Create a new test repository
func NewTestRepository(app *config.AppConfig) *Repository {
	return &Repository{ 
		App: app,
		DB: dbrepository.NewTestRepository(app),
	}
}

// Sets the repository for the handlers
func SetRepository(repo *Repository) {
	Repo = repo
}

// Send back error message as JSON
func SendJsonErrorResponse(w http.ResponseWriter, available bool, message string) {
	res := jsonResponse{
		OK: available,
		Message: message,
	}

	// Send JSON error response
	out, _ := json.MarshalIndent(res, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
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
	// Get reservation from `Session` object
	reservation, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		repo.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Get room information from database
	room, err := repo.DB.GetRoomByID(reservation.RoomID)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't find room with given id")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Store room name in `reservation` and update `reservation` in the `Session` object
	reservation.Room.RoomName = room.RoomName
	repo.App.Session.Put(r.Context(), "reservation", reservation)

	// Parse dates into YYYY-MM-DD string format
	startDate := reservation.StartDate.Format("2006-01-02")
	endDate := reservation.EndDate.Format("2006-01-02")

	// Store dates in string map
	stringMap := make(map[string]string)
	stringMap["start_date"] = startDate
	stringMap["end_date"] = endDate

	// Store reservation in data map
	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
		StringMap: stringMap,
	})
}

// Handles the creation of a reservation
func (repo *Repository) PostMakeReservation(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Parse start and end dates received as form data
	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")
	startDate, endDate, err := helpers.ParseDates(w, sd, ed)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't parse dates")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Extract room id and parse it into an integer
	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Invalid room id")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Create reservation with the form data
	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Phone:     r.Form.Get("phone"),
		Email:     r.Form.Get("email"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    roomID,
	}

	// Validate form data and add any errors that might exist to `form` variable
	form := forms.New(r.PostForm)
	form.RequiredFields("first_name", "last_name", "email")
	form.MinLength("first_name", 2)
	form.IsEmail("email")

	// Rerender make reservation form with updated error information
	if !form.IsValid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		http.Error(w, "Invalid form data", http.StatusSeeOther)
		render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})

		return
	}

	// Insert reservation into database
	reservationID, err := repo.DB.InsertReservation(reservation)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't insert reservation into the database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Save data in a RoomRestriction struct and save it in the database
	roomRestriction := models.RoomRestriction{
		StartDate: startDate,
		EndDate: endDate,
		RoomID: roomID,
		ReservationID: reservationID,
		RestrictionID: 1,
	}

	err = repo.DB.InsertRoomRestriction(roomRestriction)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't insert room restriction into the database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Update `reservation` in `Session` object
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
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Parse start and end dates received as form data
	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")
	startDate, endDate, err := helpers.ParseDates(w, sd, ed)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't parse dates")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Search for availability in all rooms
	rooms, err := repo.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't get available rooms")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// If there is no availability
	if len(rooms) == 0 {
		repo.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	// Store start and end dates in the `Session` object
	// This data will be used later in the `make-reservation` page
	reservation := models.Reservation{
		StartDate: startDate,
		EndDate: endDate,
	}
	repo.App.Session.Put(r.Context(), "reservation", reservation)

	// Render `choose-room` page with available rooms information
	data := make(map[string]interface{})
	data["rooms"] = rooms

	render.RenderTemplate(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

type jsonResponse struct {
	OK 				bool 		`json:"ok"`
	Message 	string 	`json:"message"`
	RoomID 		string	`json:"room_id"`
	StartDate string 	`json:"start_date"`
	EndDate 	string 	`json:"end_date"`
}

// Accepts form data for search availability and returns a JSON response
func (repo *Repository) SearchAvailabilityJson(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		SendJsonErrorResponse(w, false, "Internal Server Error")
		return
	}
	
	// Parse start and end dates received as form data
	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")
	startDate, endDate, err := helpers.ParseDates(w, sd, ed)
	if err != nil {
		SendJsonErrorResponse(w, false, "Internal Server Error")
		return
	}

	// Extract room id and parse it into an integer
	roomId, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		SendJsonErrorResponse(w, false, "Internal Server Error")
		return
	}

	available, err := repo.DB.SearchAvailabilityByDatesAndRoom(startDate, endDate, roomId)
	if err != nil {
		SendJsonErrorResponse(w, false, "Error connecting to database")
		return
	}

	// Response to send back as JSON
	res := jsonResponse{
		OK: available,
		Message: "",
		StartDate: sd,
		EndDate: ed,
		RoomID: strconv.Itoa(roomId),
	}

	// Convert response to JSON
	jsonRes, _ := json.MarshalIndent(res, "", "    ")

	// Send back the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonRes)
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
	// Get make reservation form data from `Session` object
	reservation, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation) // Type assertion
	if !ok {
		repo.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Delete reservation data from session object
	repo.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	// Parse start and end dates into YYYY-MM-DD string format
	// Store them in a string map
	startDate := reservation.StartDate.Format("2006-01-02")
	endDate := reservation.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = startDate
	stringMap["end_date"] = endDate

	render.RenderTemplate(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
		StringMap: stringMap,
	})
}

// Handler for when user selects a room in the 'choose-room' page
// Gets the id of the selected room and stores it in the `Session` object,
// redirecting the user to the `make-reservation` page
func (repo *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	// Extract id from URL parameters by splitting the URL up by '/' and grab the 3rd element
	// Finally parse the id into an integer
	exploded := strings.Split(r.RequestURI, "/")
	if len(exploded) != 3 {
		repo.App.Session.Put(r.Context(), "error", "Missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomID, err := strconv.Atoi(exploded[2])
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Retrieve the start and end dates
	// Parse dates into the appropriate type
	reservation, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		repo.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Update reservation stored in `Session` object
	reservation.RoomID = roomID
	repo.App.Session.Put(r.Context(), "reservation", reservation)

	// Redirect user to `make-reservation` page
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// Handler to store start and end dates and room id in `Session` object
// and redirect user to `make-reservation` page in order to finish his/her booking.
// This handler is called in the modal that appears after a user checked availability for a given room
func (repo *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	// Extract room id from URL query parameters and parse it into an integer
	roomID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Missing query parameter")
		http.Redirect(w, r, "/search-availability", http.StatusTemporaryRedirect)
		return
	}

	// Extract start and end dates from URL query parameters and parse them into `date` type
	sd := r.URL.Query().Get("start_date")
	ed := r.URL.Query().Get("end_date")
	startDate, endDate, err := helpers.ParseDates(w, sd, ed)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Missing query parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Get room information from database
	room, err := repo.DB.GetRoomByID(roomID)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't get room from database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Create reservation
	var reservation models.Reservation
	reservation.RoomID = roomID
	reservation.StartDate = startDate
	reservation.EndDate = endDate
	reservation.Room.RoomName = room.RoomName

	// Set `reservation` in the `Session` object
	repo.App.Session.Put(r.Context(), "reservation", reservation)

	// Redirect user to `make-reservation` page
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}