package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/config"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/driver"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/forms"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/helpers"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/render"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/repository"
	dbrepository "github.com/LuisBarroso37/bed-and-breakfast/internal/repository/db-repository"
	"github.com/go-chi/chi/v5"
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Get room information from database
	room, err := repo.DB.GetRoomByID(reservation.RoomID)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't find room with given id")
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Parse start and end dates received as form data
	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")
	startDate, endDate, err := helpers.ParseDates(w, sd, ed)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't parse dates")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Extract room id and parse it into an integer
	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Invalid room id")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Get room information from room id
	room, err := repo.DB.GetRoomByID(roomID)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "can't find room!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Create reservation with the form data
	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Phone:     r.Form.Get("phone"),
		Email:     r.Form.Get("email"),
		StartDate: startDate,
		Room: room,
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

		stringMap := make(map[string]string)
		stringMap["start_date"] = sd
		stringMap["end_date"] = ed

		render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
			StringMap: stringMap,
		})

		return
	}

	// Insert reservation into database
	reservationID, err := repo.DB.InsertReservation(reservation)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't insert reservation into the database")
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Send email to guest
	htmlMessage := fmt.Sprintf(`
			<strong>Reservation confirmation</strong><br>
			Dear %s, <br>
			This is to confirm your reservation of the %s from %s to %s.
		`, reservation.FirstName,
		reservation.Room.RoomName, 
		reservation.StartDate.Format("2006-01-02"), 
		reservation.EndDate.Format("2006-01-02"),
	)

	msg := models.MailData{
		To: reservation.Email,
		From: "me@here.com",
		Subject: "Reservation confirmation",
		Content: htmlMessage,
		Template: "basic.html",
	}
	repo.App.MailChan <- msg

	// Send email to owner
	htmlMessage = fmt.Sprintf(`
			<strong>Reservation confirmation</strong><br>
			Dear %s:, <br>
			This is to confirm your reservation of the %s from %s to %s.
		`, reservation.FirstName,
		reservation.Room.RoomName, 
		reservation.StartDate.Format("2006-01-02"), 
		reservation.EndDate.Format("2006-01-02"),
	)

	msg = models.MailData{
		To: "me@here.com",
		From: "me@here.com",
		Subject: "Reservation confirmation",
		Content: htmlMessage,
		Template: "basic.html",
	}
	repo.App.MailChan <- msg

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
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Parse start and end dates received as form data
	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")
	startDate, endDate, err := helpers.ParseDates(w, sd, ed)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't parse dates")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Search for availability in all rooms
	rooms, err := repo.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't get available rooms")
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	roomID, err := strconv.Atoi(exploded[2])
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Missing url parameter")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Retrieve the start and end dates
	// Parse dates into the appropriate type
	reservation, ok := repo.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		repo.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	// Extract start and end dates from URL query parameters and parse them into `date` type
	sd := r.URL.Query().Get("start_date")
	ed := r.URL.Query().Get("end_date")
	startDate, endDate, err := helpers.ParseDates(w, sd, ed)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Missing query parameter")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	// Get room information from database
	room, err := repo.DB.GetRoomByID(roomID)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't get room from database")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
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

// ShowLogin is the login page handler
func (repo *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}


// Handler to login user given a valid email and password
func (repo *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	// Prevent session fixation attack
	err := repo.App.Session.RenewToken(r.Context())
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Error renewing Session token")
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
	}

	err = r.ParseForm()
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
	}

	form := forms.New(r.PostForm)
	form.RequiredFields("email", "password")
	form.IsEmail("email")
	if !form.IsValid() {
		// Display errors in Login page
		render.RenderTemplate(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	// Extract email and password from form data
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	// Authenticate user
	id, _, err := repo.DB.Authenticate(email, password)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	// Store user id in `Session`
	repo.App.Session.Put(r.Context(), "user_id", id)
	repo.App.Session.Put(r.Context(), "success", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout user
func (repo *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	repo.App.Session.Destroy(r.Context())
	repo.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}

// AdminDahboard is the admin dashboard page handler
func (repo *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

// AdminNewReservations is the new reservations page handler in the admin dashboard
func (repo *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := repo.DB.GetNewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations
	
	render.RenderTemplate(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminAllReservations is the all-reservations page handler in the admin dashboard
func (repo *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := repo.DB.GetAllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations
	
	render.RenderTemplate(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminShowReservation is the show reservation page handler in the admin dashboard
func (repo *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	// Extract source (all or new) and id from URL
	src := chi.URLParam(r, "src")
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	currentMonth := r.URL.Query().Get("m")
	currentYear := r.URL.Query().Get("y")

	// Create string map and add it to the template
	stringMap := make(map[string]string)
	stringMap["src"] = src
	stringMap["current_month"] = currentMonth
	stringMap["current_year"] = currentYear

	// Get reservation from database
	reservation, err := repo.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Create data map and add it to the template
	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.RenderTemplate(w, r, "admin-show-reservation.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data: data,
		Form: forms.New(nil),
	})
}

// Handler to update reservation with received form data
func (repo *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request) {
	// Parse form data 
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Extract source (all or new) and id from URL
	src := chi.URLParam(r, "src")
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Create string map and add it to the template
	stringMap := make(map[string]string)
	stringMap["src"] = src

	reservation, err := repo.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Store form data in the reservation
	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	// Validate form data and add any errors that might exist to `form` variable
	form := forms.New(r.PostForm)
	form.RequiredFields("first_name", "last_name", "email", "phone")
	form.MinLength("first_name", 2)
	form.MinLength("last_name", 2)
	form.IsEmail("email")

	// Rerender make reservation form with updated error information
	if !form.IsValid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.RenderTemplate(w, r, "admin-show-reservation.page.tmpl", &models.TemplateData{
			StringMap: stringMap,
			Form: form,
			Data: data,
		})

		return
	}

	// Update reservation
	err = repo.DB.UpdateReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Store success message in `Session`
	repo.App.Session.Put(r.Context(), "success", "Reservation successfully updated")

	// Get month and year from form (if user came from reservations calendar)
	month := r.Form.Get("month")
	year := r.Form.Get("year")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/%s-reservations", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

// AdminReservationsCalendar is the reservations calendar page handler in the admin dashboard
func (repo *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	// Get current date
	currentDate := time.Now()

	// Extract year and month from request's query parameters and convert them into integers, if they exist
	if r.URL.Query().Get("y") != "" {
		year, err := strconv.Atoi(r.URL.Query().Get("y"))
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		month, err := strconv.Atoi(r.URL.Query().Get("m"))
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		// Set `currentDate` to have the given year and month
		currentDate = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	// Store current date (time.Time) in data map
	data := make(map[string]interface{})
	data["current_date"] = currentDate

	// Add 1 month to current date
	next := currentDate.AddDate(0, 1, 0)
	// Subtract 1 month to current date
	last := currentDate.AddDate(0, -1, 0)

	// Format dates
	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")
	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	// Store dates in string map
	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear
	stringMap["current_month"] = currentDate.Format("01")
	stringMap["current_month_year"] = currentDate.Format("2006")

	// Store number of days in the current month in integer map
	currentYear, currentMonth, _ := currentDate.Date()
	currentLocation := currentDate.Location()
	firstDayOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)
	
	intMap := make(map[string]int)
	intMap["days_in_month"] = lastDayOfMonth.Day()

	// Get all rooms
	rooms, err := repo.DB.GetAllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Store rooms in data map
	data["rooms"] = rooms

	for _, room := range rooms {
		// Create reservation and owner block maps
		reservationMap := make(map[string]int)
		ownerBlockMap := make(map[string]int)

		// Add an entry in the maps for every day of the month
		for day := firstDayOfMonth; !day.After(lastDayOfMonth); day = day.AddDate(0, 0, 1) {
			reservationMap[day.Format("2006-01-2")] = 0
			ownerBlockMap[day.Format("2006-01-2")] = 0
		}

		// Get all restrictions for the room in the current month
		restrictions, err := repo.DB.GetRestrictionsForRoomByDate(room.ID, firstDayOfMonth, lastDayOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		for _, restriction := range restrictions {
			// If reservation id is bigger than 0, it is a reservation
			if restriction.ReservationID > 0 {
				// Loop through the dates between the start date and the end date of the restriction
				for day := restriction.StartDate; !day.After(restriction.EndDate); day = day.AddDate(0, 0, 1) {
					reservationMap[day.Format("2006-01-2")] = restriction.ReservationID
				}
			} else {
				// Otherwise it is an owner block
				// Each block is 1 day long
				ownerBlockMap[restriction.StartDate.Format("2006-01-2")] = restriction.ID
			}
		}

		// Store maps in the data map with a reference to the room
		data[fmt.Sprintf("reservation_map_%d", room.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", room.ID)] = ownerBlockMap

		// Store owner block map in `Session`.
		// This will be used in the POST handler to compare the calendar that was first rendered 
		// with the updated calendar after the user interacts with the calendar
		repo.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", room.ID), ownerBlockMap)
	}

	render.RenderTemplate(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		IntMap: intMap,
		Data: data,
	})
}

// Handler to mark a reservation as processed
func (repo *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	// Extract source (all or new) and id from URL
	src := chi.URLParam(r, "src")
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Extract query parameters
	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	// Mark reservation as processed
	err = repo.DB.UpdateProcessedForReservation(id, true)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Store success message in `Session`
	repo.App.Session.Put(r.Context(), "success", "Reservation marked as processed")

	if year == "" && src == "calendar" {
		http.Redirect(w, r, "/admin/reservations-calendar", http.StatusSeeOther)
	} else if year == "" && src != "calendar" {
		http.Redirect(w, r, fmt.Sprintf("/admin/%s-reservations", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

// Handler to delete a reservation
func (repo *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	// Extract source (all or new) and id from URL
	src := chi.URLParam(r, "src")
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Delete the reservation
	err = repo.DB.DeleteReservation(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Extract query parameters
	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	// Store success message in `Session`
	repo.App.Session.Put(r.Context(), "success", "Reservation marked as processed")

	if year == "" && src == "calendar" {
		http.Redirect(w, r, "/admin/reservations-calendar", http.StatusSeeOther)
	} else if year == "" && src != "calendar" {
		http.Redirect(w, r, fmt.Sprintf("/admin/%s-reservations", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

// Handler to save new blocks picked by the owner (block dates so no reservations can be done for the given dates)
func (repo *Repository) AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	// Parse the form
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Get current month and year that user is in while on the reservations calendar page
	currentMonth, err := strconv.Atoi(r.Form.Get("m"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	currentYear, err := strconv.Atoi(r.Form.Get("y"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Get all rooms
	rooms, err := repo.DB.GetAllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)

	for _, room := range rooms {
		// Get owner block map from `Session`
		ownerBlockMap := repo.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", room.ID)).(map[string]int)
		// Loop through map and if we have an entry in the map that
		// does not exist in our form data and the restriction id > 0,
		// then it is a block we need to remove
		for date, value := range ownerBlockMap {
			// Only pay attention to values > 0 and that are not in the form data
			if value == 0 {
				continue
			}

			if !form.Has(fmt.Sprintf("remove_block_%d_%s", room.ID, date)) {
				// Delete the restriction by id
				err := repo.DB.DeleteBlockByID(value)
				if err != nil {
					helpers.ServerError(w, err)
					return
				}
			}
		}
	}
	
	// Loop through form data and only pay attention to properties
	// with prefix 'add_block'. These are the blocks that we have to
	// store in the database, since they are not present in the `ownerBlockMap`
	for name := range r.PostForm {
		if strings.HasPrefix(name, "add_block") {
			// Split `name` by '_' and get room id and date
			exploded := strings.Split(name, "_")

			date, err := time.Parse("2006-01-2", exploded[3])
			if err != nil {
				helpers.ServerError(w, err)
				return
			}

			roomID, err := strconv.Atoi(exploded[2])
			if err != nil {
				helpers.ServerError(w, err)
				return
			}

			// Insert owner block as room restriction for given room
			err = repo.DB.InsertBlockForRoom(roomID, date)
			if err != nil {
				helpers.ServerError(w, err)
				return
			}
		}
	}

	// Store success message in `Session` and redirect user to exact same page user was in
	repo.App.Session.Put(r.Context(), "success", "Changes saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", currentYear, currentMonth), http.StatusSeeOther)
}