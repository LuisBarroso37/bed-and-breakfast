package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/driver"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
	"github.com/go-chi/chi/v5"
)

func getRequestContext(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}

	return ctx
}

func TestNewRepository(t *testing.T) {
	var db driver.DB
	testRepo := NewRepository(&app, &db)

	// Check if variable returned by NewRepository() is of type *Repository
	testRepoType := reflect.TypeOf(testRepo).String()
	if testRepoType != "*handlers.Repository" {
		t.Errorf("Did not get correct type from NewRepository: got %s, wanted *Repository", testRepoType)
	}
}

// This contains only route handlers which do not require a `Session` object
var testsWithoutSession = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"generals-quarters", "/generals-quarters", "GET", http.StatusOK},
	{"majors-suite", "/majors-suite", "GET", http.StatusOK},
	{"search-availability", "/search-availability", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
	{"non-existent route", "/invalid-route", "GET", http.StatusNotFound},
	{"login", "/auth/login", "GET", http.StatusOK},
	{"logout", "/auth/logout", "GET", http.StatusOK},
	{"admin dashboard", "/admin/dashboard", "GET", http.StatusOK},
	{"admin all reservations", "/admin/all-reservations", "GET", http.StatusOK},
	{"admin new reservations", "/admin/new-reservations", "GET", http.StatusOK},
	{"admin show reservation", "/admin/reservations/new/1", "GET", http.StatusOK},
	{"admin resservation calendar", "/admin/reservations-calendar", "GET", http.StatusOK},
	{"admin resservation calendar with query params", "/admin/reservations-calendar?y=2020&m=1", "GET", http.StatusOK},
}

func TestHandlersThatDoNotRequireSession(t *testing.T) {
	// Setup routes
	routes := getRoutes()

	// Create test server
	testServer := httptest.NewTLSServer(routes)
	defer testServer.Close()

	for _, test := range testsWithoutSession {
		res, err := testServer.Client().Get(testServer.URL + test.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if res.StatusCode != test.expectedStatusCode {
			t.Errorf("for %s, expected %d but got %d", test.name, test.expectedStatusCode, res.StatusCode)
		}
	}
}

var makeReservationTests = []struct {
	name               	string
	reservation				 	models.Reservation
	shouldStoreSession	bool
	expectedStatusCode 	int
	expectedRedirectURL string
	expectedHTML 				string
}{
	{
		"Success",
		models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID: 1,
				RoomName: "General's Quarters",
			},
		}, 
		true,
		http.StatusOK,
		"",
		`action="/make-reservation"`,
	},
	{
		"Reservation not in session", 
		models.Reservation{}, 
		false,
		http.StatusSeeOther,
		"/",
		"",
	},
	{
		"Room does not exist", 
		models.Reservation{
			RoomID: 3,
			Room: models.Room{
				ID: 3,
				RoomName: "General's Quarters",
			},
		}, 
		true, 
		http.StatusSeeOther,
		"/",
		"",
	},
}

func TestRepository_MakeReservation(t *testing.T) {
	for _, test := range makeReservationTests {
		// Create dummy reservation
		reservation := test.reservation

		// Create http request to `/make-reservation` 
		// and store context on it which includes the `X-Session` header
		// in order to read to/from the `Session object`
		req, err := http.NewRequest("GET", "/make-reservation", nil)
		if err != nil {
			log.Println(err)
		}
		ctx := getRequestContext(req)
		req = req.WithContext(ctx)

		// This fakes all of the request/response lifecycle
		// Stores the response we get from the request
		responseRecorder := httptest.NewRecorder()

		// Store dummy reservation in the `Session` object if `shouldStoreSession` flag is true
		if test.shouldStoreSession {
			session.Put(ctx, "reservation", reservation)
		}

		// Make handler function able to be called directly and execute it
		handler := http.HandlerFunc(Repo.MakeReservation)
		handler.ServeHTTP(responseRecorder, req)

		if responseRecorder.Code != test.expectedStatusCode {
			t.Errorf(
				"Test '%s' returns wrong response status code: got %d, wanted %d",
				test.name,
				responseRecorder.Code,
				test.expectedStatusCode,
			)
		}

		if test.expectedRedirectURL != "" {
			// Get redirect URL
			redirectURL, err := responseRecorder.Result().Location()
			if err != nil {
				log.Println(err)
			}

			if redirectURL.String() != test.expectedRedirectURL {
				t.Errorf(
					"Test %s redirects user to wrong URL: got %s, wanted %s",
					test.name,
					redirectURL.String(),
					test.expectedRedirectURL,
				)
			}
		}

		if test.expectedHTML != "" {
			html := responseRecorder.Body.String()

			if !strings.Contains(html, test.expectedHTML) {
				t.Errorf(
					"Test %s return wrong HTML: expected %s",
					test.name,
					html,
				)
			}
		}
	}
}

var postMakeReservationTests = []struct {
	name               	string
	body							 	url.Values
	expectedStatusCode 	int
	expectedRedirectURL string
	expectedHTML				string
}{
	{
		"Success", 
		url.Values{
			"start_date": []string{"2050-01-01"},
			"end_date": []string{"2050-01-02"},
			"first_name": []string{"John"},
			"last_name": []string{"Smith"},
			"email": []string{"john@smith.com"},
			"phone": []string{"123456789"},
			"room_id": []string{"1"},
		}, 
		http.StatusSeeOther,
		"/reservation-summary",
		"",
	},
	{
		"Unable to parse form", 
		nil, 
		http.StatusSeeOther,
		"/",
		"",
	},
	{
		"Invalid start date", 
		url.Values{
			"start_date": []string{"invalid"},
			"end_date": []string{"2050-01-02"},
			"first_name": []string{"John"},
			"last_name": []string{"Smith"},
			"email": []string{"john@smith.com"},
			"phone": []string{"123456789"},
			"room_id": []string{"1"},
		}, 
		http.StatusSeeOther,
		"/",
		"",
	},
	{
		"Invalid end date", 
		url.Values{
			"start_date": []string{"2050-01-01"},
			"end_date": []string{"invalid"},
			"first_name": []string{"John"},
			"last_name": []string{"Smith"},
			"email": []string{"john@smith.com"},
			"phone": []string{"123456789"},
			"room_id": []string{"1"},
		}, 
		http.StatusSeeOther,
		"/",
		"",
	},
	{
		"Invalid room id", 
		url.Values{
			"start_date": []string{"2050-01-01"},
			"end_date": []string{"2050-01-02"},
			"first_name": []string{"John"},
			"last_name": []string{"Smith"},
			"email": []string{"john@smith.com"},
			"phone": []string{"123456789"},
			"room_id": []string{"invalid"},
		}, 
		http.StatusSeeOther,
		"/",
		"",
	},
	{
		"Invalid form data", 
		url.Values{
			"start_date": []string{"2050-01-01"},
			"end_date": []string{"2050-01-02"},
			"first_name": []string{"L"},
			"last_name": []string{"Xiang"},
			"email": []string{"l@xiang.com"},
			"phone": []string{"123456789"},
			"room_id": []string{"1"},
		}, 
		http.StatusOK,
		"",
		`action="/make-reservation"`,
	},
	{
		"Failure to insert reservation in database", 
		url.Values{
			"start_date": []string{"2050-01-01"},
			"end_date": []string{"2050-01-02"},
			"first_name": []string{"John"},
			"last_name": []string{"Smith"},
			"email": []string{"john@smith.com"},
			"phone": []string{"123456789"},
			"room_id": []string{"2"},
		}, 
		http.StatusSeeOther,
		"/",
		"",
	},
	{
		"Failure to insert room restriction in database", 
		url.Values{
			"start_date": []string{"2050-01-01"},
			"end_date": []string{"2050-01-02"},
			"first_name": []string{"John"},
			"last_name": []string{"Smith"},
			"email": []string{"john@smith.com"},
			"phone": []string{"123456789"},
			"room_id": []string{"1000"},
		}, 
		http.StatusSeeOther,
		"/",
		"",
	},
}

func TestRepository_PostMakeReservation(t *testing.T) {
	for _, test := range postMakeReservationTests {
		// Set request body depending on if it is a POST or a GET request
		var reqBody io.Reader
		if test.body == nil {
			reqBody = nil
		} else {
			reqBody = strings.NewReader(test.body.Encode())
		}

		// Create http request to `/make-reservation` 
		// and store context on it which includes the `X-Session` header
		// in order to read to/from the `Session object`
		req, err := http.NewRequest("POST", "/make-reservation", reqBody)
		if err != nil {
			log.Println(err)
		}
		ctx := getRequestContext(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// This fakes all of the request/response lifecycle
		// Stores the response we get from the request
		responseRecorder := httptest.NewRecorder()

		// Make handler function able to be called directly and execute it
		handler := http.HandlerFunc(Repo.PostMakeReservation)
		handler.ServeHTTP(responseRecorder, req)

		if responseRecorder.Code != test.expectedStatusCode {
			t.Errorf(
				"Test '%s' returns wrong response status code: got %d, wanted %d",
				test.name,
				responseRecorder.Code,
				test.expectedStatusCode,
			)
		}

		if test.expectedRedirectURL != "" {
			// Get redirect URL
			redirectURL, err := responseRecorder.Result().Location()
			if err != nil {
				log.Println(err)
			}

			if redirectURL.String() != test.expectedRedirectURL {
				t.Errorf(
					"Test %s redirects user to wrong URL: got %s, wanted %s",
					test.name,
					redirectURL.String(),
					test.expectedRedirectURL,
				)
			}
		}

		if test.expectedHTML != "" {
			html := responseRecorder.Body.String()

			if !strings.Contains(html, test.expectedHTML) {
				t.Errorf(
					"Test %s return wrong HTML: expected %s",
					test.name,
					html,
				)
			}
		}
	}
}

var postSearchAvailabilityTests = []struct {
	name               	string
	body         				url.Values
	expectedStatusCode 	int
}{
	{
		"Rooms not available",
		url.Values{
			"start_date": {"2050-01-01"},
			"end_date": {"2050-01-02"},
		},
		http.StatusSeeOther,
	},
	{
		"Rooms are available",
		url.Values{
			"start_date": {"2049-01-01"},
			"end_date": {"2049-01-02"},
			"room_id": {"1"},
		},
		http.StatusOK,
	},
	{
		"Empty request body",
		url.Values{},
		http.StatusSeeOther,
	},
	{
		"Invalid start date",
		url.Values{
			"start_date": {"invalid"},
			"end_date": {"2040-01-02"},
			"room_id": {"1"},
		},
		http.StatusSeeOther,
	},
	{
		"Invalid end date",
		url.Values{
			"start_date": {"2040-01-01"},
			"end_date": {"invalid"},
		},
		http.StatusSeeOther,
	},
	{
		"Database query fails",
		url.Values{
			"start_date": {"2000-01-01"},
			"end_date": {"2000-01-02"},
		},
		http.StatusSeeOther,
	},
}


func TestRepository_PostSearchAvailability(t *testing.T) {
	for _, test := range postSearchAvailabilityTests {
		// Create POST request to `/search-availability` 
		// and store context on it which includes the `X-Session` header
		// in order to read to/from the `Session object`
		req, err := http.NewRequest("POST", "/search-availability", strings.NewReader(test.body.Encode()))
		if err != nil {
			log.Println(err)
		}
		ctx := getRequestContext(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// This fakes all of the request/response lifecycle
		// Stores the response we get from the request
		responseRecorder := httptest.NewRecorder()

		// Make `search-availability` handler function able to be called directly
		// and execute it
		handler := http.HandlerFunc(Repo.PostSearchAvailability)
		handler.ServeHTTP(responseRecorder, req)

		if responseRecorder.Code != test.expectedStatusCode {
			t.Errorf(
				"Test %s returns wrong response status code: got %d, wanted %d",
				test.name,
				responseRecorder.Code,
				test.expectedStatusCode,
			)
		}
	}
}

var searchAvailabilityJsonTests = []struct {
	name            string
	body      			url.Values
	expectedOK      bool
	expectedMessage string
}{
	{
		"Rooms not available",
		url.Values{
			"start_date": {"2050-01-01"},
			"end_date": {"2050-01-02"},
			"room_id": {"1"},
		},
		false,
		"",
	}, {
		"Rooms are available",
		url.Values{
			"start_date": {"2049-01-01"},
			"end_date": {"2049-01-02"},
			"room_id": {"1"},
		},
		true,
		"",
	},
	{
		"Empty request body",
		nil,
		false,
		"Internal Server Error",
	},
	{
		"Invalid start date",
		url.Values{
			"start_date": {"invalid"},
			"end_date": {"2049-01-02"},
			"room_id": {"1"},
		},
		false,
		"Internal Server Error",
	},
	{
		"Invalid end date",
		url.Values{
			"start_date": {"2049-01-01"},
			"end_date": {"invalid"},
			"room_id": {"1"},
		},
		false,
		"Internal Server Error",
	},
	{
		"Invalid room id",
		url.Values{
			"start_date": {"2049-01-01"},
			"end_date": {"2049-01-02"},
			"room_id": {"invalid"},
		},
		false,
		"Internal Server Error",
	},
	{
		"Database query fails",
		url.Values{
			"start_date": {"2000-01-01"},
			"end_date": {"2000-01-02"},
			"room_id": {"1"},
		},
		false,
		"Error connecting to database",
	},
}

func TestAvailabilityJSON(t *testing.T) {
	for _, test := range searchAvailabilityJsonTests {
		// Set request body depending on if it is a POST or a GET request
		var reqBody io.Reader
		if test.body == nil {
			reqBody = nil
		} else {
			reqBody = strings.NewReader(test.body.Encode())
		}
		
		// Create POST request to `/search-availability-json` 
		// and store context on it which includes the `X-Session` header
		// in order to read to/from the `Session object`
		req, err := http.NewRequest("POST", "/search-availability-json", reqBody)
		if err != nil {
			log.Println(err)
		}

		ctx := getRequestContext(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// This fakes all of the request/response lifecycle
		// Stores the response we get from the request
		responseRecorder := httptest.NewRecorder()

		// Make `search-availability` handler function able to be called directly
		// and execute it
		handler := http.HandlerFunc(Repo.SearchAvailabilityJson)
		handler.ServeHTTP(responseRecorder, req)

		var jsonRes jsonResponse

		err = json.Unmarshal(responseRecorder.Body.Bytes(), &jsonRes)
		if err != nil {
			t.Error("Failed to parse json response")
		}

		if jsonRes.OK != test.expectedOK {
			t.Errorf("Test %s failed: expected OK to be %v but got %v", test.name, test.expectedOK, jsonRes.OK)
		}

		if jsonRes.Message != test.expectedMessage  {
			t.Errorf("Test %s failed: expected message to be %s but got %s", test.name, test.expectedMessage, jsonRes.Message)
		}
	}
}

var reservationSummaryTests = []struct {
	name               		string
	reservation        		models.Reservation
	expectedStatusCode 		int
	expectedRedirectURL   string
}{
	{
		"Reservation in session",
		models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		http.StatusOK,
		"",
	},
	{
		"No reservation in session",
		models.Reservation{},
		http.StatusSeeOther,
		"/",
	},
}


func TestRepository_ReservationSummary_Success(t *testing.T) {
	for _, test := range reservationSummaryTests {
		// Create http request to `/reservation-summary` 
		// and store context on it which includes the `X-Session` header
		// in order to read to/from the `Session object`
		req, err := http.NewRequest("GET", "/reservation-summary", nil)
		if err != nil {
			log.Println(err)
		}
		ctx := getRequestContext(req)
		req = req.WithContext(ctx)

		// This fakes all of the request/response lifecycle
		// Stores the response we get from the request
		responseRecorder := httptest.NewRecorder()

		// Store dummy reservation in the `Session` object if reservation is not empty
		if test.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", test.reservation)
		}

		// Make `reservation-summary` handler function able to be called directly
		// and execute it
		handler := http.HandlerFunc(Repo.ReservationSummary)
		handler.ServeHTTP(responseRecorder, req)

		if responseRecorder.Code != test.expectedStatusCode {
			t.Errorf(
				"Test %s returns wrong response status code: got %d, wanted %d",
				test.name,
				responseRecorder.Code,
				test.expectedStatusCode,
			)
		}
	}
}

var chooseRoomTests = []struct {
	name               		string
	reservation        		models.Reservation
	url                		string
	expectedStatusCode 		int
	expectedRedirectURL   string
}{
	{
		"Reservation in session",
		models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		"/choose-room/1",
		http.StatusSeeOther,
		"/make-reservation",
	},
	{
		"Reservation not in session",
		models.Reservation{},
		"/choose-room/1",
		http.StatusSeeOther,
		"/",
	},
	{
		"Missing url parameter",
		models.Reservation{},
		"/choose-room",
		http.StatusSeeOther,
		"/",
	},
	{
		"Invalid url parameter",
		models.Reservation{},
		"/choose-room/invalid",
		http.StatusSeeOther,
		"/",
	},
}

func TestRepository_ChooseRoom(t *testing.T) {
	for _, test := range chooseRoomTests {
		// Create http request to `/choose-room/1` 
		// and store context on it which includes the `X-Session` header
		// in order to read to/from the `Session object`
		req, err := http.NewRequest("GET", test.url, nil)
		if err != nil {
			log.Println(err)
		}
		ctx := getRequestContext(req)
		req = req.WithContext(ctx)
		
		// Set the RequestURI on the request so that we can grab the ID
		// from the URL
		req.RequestURI = test.url

		// This fakes all of the request/response lifecycle
		// Stores the response we get from the request
		responseRecorder := httptest.NewRecorder()

		// Store dummy reservation in the `Session` object
		if test.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", test.reservation)
		}

		// Make `choose-room` handler function able to be called directly
		// and execute it
		handler := http.HandlerFunc(Repo.ChooseRoom)
		handler.ServeHTTP(responseRecorder, req)

		if responseRecorder.Code != test.expectedStatusCode {
			t.Errorf(
				"Test %s returns wrong response status code: got %d, wanted %d",
				test.name,
				responseRecorder.Code,
				test.expectedStatusCode,
			)
		}

		if test.expectedRedirectURL != "" {
			// Get redirect URL
			redirectURL, err := responseRecorder.Result().Location()
			if err != nil {
				log.Println(err)
			}

			if redirectURL.String() != test.expectedRedirectURL {
				t.Errorf(
					"Test %s redirects user to wrong URL: got %s, wanted %s",
					test.name,
					redirectURL.String(),
					test.expectedRedirectURL,
				)
			}
		}
	}
}

var bookRoomTests = []struct {
	name               	string
	url                	string
	expectedStatusCode 	int
	expectedRedirectURL string
}{
	{
		"Fetches room information",
		"/book-room?start_date=2049-01-01&end_date=2049-01-02&id=1",
		http.StatusSeeOther,
		"/make-reservation",
	},
	{
		"Missing id query parameter",
		"/book-room?start_date=2049-01-01&end_date=2049-01-02",
		http.StatusSeeOther,
		"/search-availability",
	},
	{
		"Invalid id query parameter",
		"/book-room?start_date=2049-01-01&end_date=2049-01-02&id=invalid",
		http.StatusSeeOther,
		"/search-availability",
	},
	{
		"Missing start_date query parameter",
		"/book-room?end_date=2049-01-02&id=1",
		http.StatusSeeOther,
		"/search-availability",
	},
	{
		"Invalid start_date query parameter",
		"/book-room?start_date=invalid&end_date=2049-01-02&id=1",
		http.StatusSeeOther,
		"/search-availability",
	},
	{
		"Missing end_date query parameter",
		"/book-room?start_date=2049-01-01&id=1",
		http.StatusSeeOther,
		"/search-availability",
	},
	{
		"Invalid end_date query parameter",
		"/book-room?start_date=2049-01-01&end_date=invalid&id=1",
		http.StatusSeeOther,
		"/search-availability",
	},
	{
		"Database query fails",
		"/book-room?start_date=2049-01-01&end_date=2049-01-02&id=3",
		http.StatusSeeOther,
		"/search-availability",
	},
}

func TestRepository_BookRoom(t *testing.T) {
	for _, test := range bookRoomTests {
		// Create http request to `/book-room` 
		// and store context on it which includes the `X-Session` header
		// in order to read to/from the `Session object`
		req, err := http.NewRequest("GET", test.url, nil)
		if err != nil {
			log.Println(err)
		}
		ctx := getRequestContext(req)
		req = req.WithContext(ctx)

		// This fakes all of the request/response lifecycle
		// Stores the response we get from the request
		responseRecorder := httptest.NewRecorder()

		// Make `book-room` handler function able to be called directly
		// and execute it
		handler := http.HandlerFunc(Repo.BookRoom)
		handler.ServeHTTP(responseRecorder, req)

		if responseRecorder.Code != test.expectedStatusCode {
			t.Errorf(
				"Test %s returns wrong response status code: got %d, wanted %d",
				test.name,
				responseRecorder.Code,
				test.expectedStatusCode,
			)
		}

		if test.expectedRedirectURL != "" {
			// Get redirect URL
			redirectURL, err := responseRecorder.Result().Location()
			if err != nil {
				log.Println(err)
			}

			if redirectURL.String() != test.expectedRedirectURL {
				t.Errorf(
					"Test %s redirects user to wrong URL: got %s, wanted %s",
					test.name,
					redirectURL.String(),
					test.expectedRedirectURL,
				)
			}
		}
	}
}

var postShowLoginTests = []struct {
	name 								string
	body 								url.Values
	expectedStatusCode 	int
	expectedHTML 				string
	expectedRedirectURL string
} {
	{
		"Valid credentials", 
		url.Values{
			"email": {"me@here.com"},
			"password": {"password"},
		}, 
		http.StatusSeeOther, 
		"", 
		"/",
	},
	{
		"Invalid credentials",
		url.Values{
			"email": {"invalid@here.com"},
			"password": {"password"},
		}, 
		http.StatusSeeOther, 
		"", 
		"/auth/login",
	},
	{
		"Empty request body", 
		nil, 
		http.StatusSeeOther, 
		"", 
		"/auth/login",
	},
	{
		"Missing email", 
		url.Values{
			"password": {"password"},
		}, 
		http.StatusOK, 
		`action="/auth/login"`, 
		"",
	},
	{
		"Invalid email",
		url.Values{
			"email": {"invalid@"},
			"password": {"password"},
		}, 
		http.StatusOK, 
		`action="/auth/login"`, 
		"",
	},
	{
		"Missing password",
		url.Values{
			"email": {"me@here.com"},
		}, 
		http.StatusOK, 
		`action="/auth/login"`, 
		"",
	},
}

func TestRepository_PostShowLogin(t *testing.T) {
	for _, test := range postShowLoginTests {
		var reqBody io.Reader

		if test.body == nil {
			reqBody = nil
		} else {
			reqBody = strings.NewReader(test.body.Encode())
		}

		// Create POST request to `/auth/login`
		// and store context on it which includes the `X-Session` header
		// in order to read to/from the `Session object`
		req, err := http.NewRequest("POST", "/auth/login", reqBody)
		if err != nil {
			log.Println(err)
		}

		ctx := getRequestContext(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// This fakes all of the request/response lifecycle
		// Stores the response we get from the request
		responseRecorder := httptest.NewRecorder()

		// Make `/auth/login` POST handler function able to be called directly
		// and execute it
		handler := http.HandlerFunc(Repo.PostShowLogin)
		handler.ServeHTTP(responseRecorder, req)

		if responseRecorder.Code != test.expectedStatusCode {
			t.Errorf(
				"Test %s returns wrong response status code: got %d, wanted %d",
				test.name,
				responseRecorder.Code,
				test.expectedStatusCode,
			)
		}

		if test.expectedRedirectURL != "" {
			// Get redirect URL
			redirectURL, err := responseRecorder.Result().Location()
			if err != nil {
				log.Println(err)
			}

			if redirectURL.String() != test.expectedRedirectURL {
				t.Errorf(
					"Test %s redirects user to wrong URL: got %s, wanted %s",
					test.name,
					redirectURL.String(),
					test.expectedRedirectURL,
				)
			}
		}

		if test.expectedHTML != "" {
			html := responseRecorder.Body.String()

			if !strings.Contains(html, test.expectedHTML) {
				t.Errorf(
					"Test %s return wrong HTML: expected %s",
					test.name,
					html,
				)
			}
		}
	}
}

var adminPostShowReservationTests = []struct {
	name                 	string
	url                  	string
	src										string
	id										string
	body           				url.Values
	expectedStatusCode 		int
	expectedRedirectURL   string
	expectedHTML				  string
}{
	{
		"Valid data coming from new-reservations page",
		"/admin/reservations/new/1",
		"new",
		"1",
		url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		http.StatusSeeOther,
		"/admin/new-reservations",
		"",
	},
	{
		"Valid data coming from all-reservations page",
		"/admin/reservations/all/1",
		"all",
		"1",
		url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		http.StatusSeeOther,
		"/admin/all-reservations",
		"",
	},
	{
		"Valid data coming from reservations-calendar page",
		"/admin/reservations/calendar/1",
		"calendar",
		"1",
		url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
			"year":       {"2022"},
			"month":      {"01"},
		},
		http.StatusSeeOther,
		"/admin/reservations-calendar?y=2022&m=01",
		"",
	},
	{
		"Empty request body",
		"/admin/reservations/new/1",
		"new",
		"1",
		nil,
		http.StatusInternalServerError,
		"",
		"",
	},
	{
		"Invalid id URL parameter",
		"/admin/reservations/all/invalid",
		"all",
		"invalid",
		url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		http.StatusInternalServerError,
		"",
		"",
	},
	{
		"Reservation id not found",
		"/admin/reservations/all/11",
		"all",
		"11",
		url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		http.StatusInternalServerError,
		"",
		"",
	},
	{
		"Reservation id not found",
		"/admin/reservations/all/11",
		"all",
		"11",
		url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		http.StatusInternalServerError,
		"",
		"",
	},
	{
		"First name length is too short",
		"/admin/reservations/all/1",
		"all",
		"1",
		url.Values{
			"first_name": {"J"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		http.StatusOK,
		"",
		`action="/admin/reservations/all/0"`,
	},
	{
		"Last name length is too short",
		"/admin/reservations/all/1",
		"all",
		"1",
		url.Values{
			"first_name": {"John"},
			"last_name":  {"S"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		http.StatusOK,
		"",
		`action="/admin/reservations/all/0"`,
	},
	{
		"Invalid email",
		"/admin/reservations/all/1",
		"all",
		"1",
		url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@"},
			"phone":      {"555-555-5555"},
		},
		http.StatusOK,
		"",
		`action="/admin/reservations/all/0"`,
	},
	{
		"Missing email",
		"/admin/reservations/all/1",
		"all",
		"1",
		url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"phone":      {"555-555-5555"},
		},
		http.StatusOK,
		"",
		`action="/admin/reservations/all/0"`,
	},
	{
		"Missing first name",
		"/admin/reservations/all/1",
		"all",
		"1",
		url.Values{
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		http.StatusOK,
		"",
		`action="/admin/reservations/all/0"`,
	},
	{
		"Missing last name",
		"/admin/reservations/all/1",
		"all",
		"1",
		url.Values{
			"first_name": {"John"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		http.StatusOK,
		"",
		`action="/admin/reservations/all/0"`,
	},
	{
		"Missing phone",
		"/admin/reservations/all/1",
		"all",
		"1",
		url.Values{
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
		},
		http.StatusOK,
		"",
		`action="/admin/reservations/all/0"`,
	},
	{
		"Reservation update failed",
		"/admin/reservations/all/1",
		"all",
		"1",
		url.Values{
			"first_name": {"Invalid"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
		},
		http.StatusInternalServerError,
		"",
		"",
	},
}

func TestRepository_AdminPostShowReservation(t *testing.T) {
	for _, test := range adminPostShowReservationTests {
		var reqBody io.Reader

		if test.body == nil {
			reqBody = nil
		} else {
			reqBody = strings.NewReader(test.body.Encode())
		}

		// Create POST request to `/admin/reservations/{src}/{id}`
		// and store context on it which includes the `X-Session` header
		// in order to read to/from the `Session object`
		req, err := http.NewRequest("POST", test.url, reqBody)
		if err != nil {
			log.Println(err)
		}

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("src", test.src)
		rctx.URLParams.Add("id", test.id)

		ctx := getRequestContext(req)
		req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// This fakes all of the request/response lifecycle
		// Stores the response we get from the request
		responseRecorder := httptest.NewRecorder()

		// Make `/admin/reservations/{src}/{id}` POST handler function able to be called directly
		// and execute it
		handler := http.HandlerFunc(Repo.AdminPostShowReservation)
		handler.ServeHTTP(responseRecorder, req)

		if responseRecorder.Code != test.expectedStatusCode {
			t.Errorf(
				"Test %s returns wrong response status code: got %d, wanted %d",
				test.name,
				responseRecorder.Code,
				test.expectedStatusCode,
			)
		}

		if test.expectedRedirectURL != "" {
			// Get redirect URL
			redirectURL, err := responseRecorder.Result().Location()
			if err != nil {
				log.Println(err)
			}

			if redirectURL.String() != test.expectedRedirectURL {
				t.Errorf(
					"Test %s redirects user to wrong URL: got %s, wanted %s",
					test.name,
					redirectURL.String(),
					test.expectedRedirectURL,
				)
			}
		}

		if test.expectedHTML != "" {
			html := responseRecorder.Body.String()

			if !strings.Contains(html, test.expectedHTML) {
				t.Errorf(
					"Test %s return wrong HTML: expected %s",
					test.name,
					html,
				)
			}
		}
	}
}

var adminPostReservationsCalendarTests = []struct {
	name                string
	body           			url.Values
	expectedStatusCode 	int
	expectedRedirectURL string
	expectedHTML        string
	blocks              int
	reservations        int
}{
	{
		name: "Add owner block as room restriction",
		body: url.Values{
			"y":  {time.Now().Format("2006")},
			"m": {time.Now().Format("01")},
			fmt.Sprintf("add_block_1_%s", time.Now().AddDate(0, 0, 2).Format("2006-01-2")): {"1"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedRedirectURL: fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", time.Now().Format("2006"), time.Now().Format("01")),
	},
	{
		name: "Empty request body",
		body: nil,
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name: "Invalid query parameter y",
		body: url.Values{
			"y":  {"invalid"},
			"m": {time.Now().Format("01")},
		},
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name: "Invalid query parameter m",
		body: url.Values{
			"y":  {time.Now().Format("2006")},
			"m": {"invalid"},
		},
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name: "Invalid date in add_block_{id}_{date}",
		body: url.Values{
			"y":  {time.Now().Format("2006")},
			"m": {time.Now().Format("01")},
			fmt.Sprintf("add_block_1_%s", "invalid"): {"1"},
		},
		expectedStatusCode: http.StatusInternalServerError,
	},
	{
		name: "Invalid date in add_block_{id}_{date}",
		body: url.Values{
			"y":  {time.Now().Format("2006")},
			"m": {time.Now().Format("01")},
			fmt.Sprintf("add_block_invalid_%s", time.Now().AddDate(0, 0, 2).Format("2006-01-2")): {"1"},
		},
		expectedStatusCode: http.StatusInternalServerError,
	},
}

func TestRepository_AdminPostReservationsCalendar(t *testing.T) {
	for _, test := range adminPostReservationsCalendarTests {
		var reqBody io.Reader

		if test.body == nil {
			reqBody = nil
		} else {
			reqBody = strings.NewReader(test.body.Encode())
		}

		// Create POST request to `/admin/reservations-calendar`
		// and store context on it which includes the `X-Session` header
		// in order to read to/from the `Session object`
		req, err := http.NewRequest("POST", "/admin/reservations-calendar", reqBody)
		if err != nil {
			log.Println(err)
		}

		// Store `Session` context in request
		ctx := getRequestContext(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Store reservation and owner block maps in `Session`
		// These will simulate the original maps when the user
		// first navigates to `/admin/reservations-calendar`
		currentDate := time.Now()
		blockMap := make(map[string]int)
		reservationMap := make(map[string]int)

		currentYear, currentMonth, _ := currentDate.Date()
		currentLocation := currentDate.Location()

		firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

		for day := firstOfMonth; day.After(lastOfMonth) == false; day = day.AddDate(0, 0, 1) {
			reservationMap[day.Format("2006-01-2")] = 0
			blockMap[day.Format("2006-01-2")] = 0
		}

		if test.blocks > 0 {
			blockMap[currentDate.Format("2006-01-2")] = test.blocks
		}

		if test.reservations > 0 {
			reservationMap[currentDate.AddDate(0, 0, 3).Format("2006-01-2")] = test.reservations
		}

		session.Put(ctx, "block_map_1", blockMap)
		session.Put(ctx, "reservation_map_1", reservationMap)

		// This fakes all of the request/response lifecycle
		// Stores the response we get from the request
		responseRecorder := httptest.NewRecorder()

		// Make `/admin/reservations-calendar` POST handler function able to be called directly
		// and execute it
		handler := http.HandlerFunc(Repo.AdminPostReservationsCalendar)
		handler.ServeHTTP(responseRecorder, req)

		if responseRecorder.Code != test.expectedStatusCode {
			t.Errorf(
				"Test %s returns wrong response status code: got %d, wanted %d",
				test.name,
				responseRecorder.Code,
				test.expectedStatusCode,
			)
		}

		if test.expectedRedirectURL != "" {
			// Get redirect URL
			redirectURL, err := responseRecorder.Result().Location()
			if err != nil {
				log.Println(err)
			}

			if redirectURL.String() != test.expectedRedirectURL {
				t.Errorf(
					"Test %s redirects user to wrong URL: got %s, wanted %s",
					test.name,
					redirectURL.String(),
					test.expectedRedirectURL,
				)
			}
		}

		if test.expectedHTML != "" {
			html := responseRecorder.Body.String()

			if !strings.Contains(html, test.expectedHTML) {
				t.Errorf(
					"Test %s return wrong HTML: expected %s",
					test.name,
					html,
				)
			}
		}
	}
}

var adminProcessReservationTests = []struct {
	name                 string
	queryParams          string
	id									 string
	src									 string
	expectedStatusCode 	 int
	expectedRedirectURL  string
}{
	{
		"Marks reservation as processed",
		"",
		"1",
		"calendar",
		http.StatusSeeOther,
		"/admin/reservations-calendar",
	},
	{
		"Marks reservation as processed and navigates user back to appropriate month and year in reservations calendar",
		"?y=2021&m=12",
		"1",
		"calendar",
		http.StatusSeeOther,
		"/admin/reservations-calendar?y=2021&m=12",
	},
	{
		"Marks reservation as processed and redirects user back to all-reservations page",
		"",
		"1",
		"all",
		http.StatusSeeOther,
		"/admin/all-reservations",
	},
	{
		"Invalid id URL parameter",
		"",
		"invalid",
		"calendar",
		http.StatusInternalServerError,
		"",
	},
	{
		"Reservation update failed",
		"",
		"11",
		"calendar",
		http.StatusInternalServerError,
		"",
	},
}

func TestRepository_AdminProcessReservation(t *testing.T) {
	for _, test := range adminProcessReservationTests {
		// Create GET request to `/admin/process-reservation/{src}/{id}`
		// and store context on it which includes the `X-Session` header
		// in order to read to/from the `Session object`
		req, err := http.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/%s/%s%s", test.src, test.id, test.queryParams), nil)
		if err != nil {
			log.Println(err)
		}

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("src", test.src)
		rctx.URLParams.Add("id", test.id)

		ctx := getRequestContext(req)
		req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// This fakes all of the request/response lifecycle
		// Stores the response we get from the request
		responseRecorder := httptest.NewRecorder()

		// Make `/admin/process-reservation/{src}/{id}` POST handler function able to be called directly
		// and execute it
		handler := http.HandlerFunc(Repo.AdminProcessReservation)
		handler.ServeHTTP(responseRecorder, req)

		if responseRecorder.Code != test.expectedStatusCode {
			t.Errorf(
				"Test %s returns wrong response status code: got %d, wanted %d",
				test.name,
				responseRecorder.Code,
				test.expectedStatusCode,
			)
		}

		if test.expectedRedirectURL != "" {
			// Get redirect URL
			redirectURL, err := responseRecorder.Result().Location()
			if err != nil {
				log.Println(err)
			}

			if redirectURL.String() != test.expectedRedirectURL {
				t.Errorf(
					"Test %s redirects user to wrong URL: got %s, wanted %s",
					test.name,
					redirectURL.String(),
					test.expectedRedirectURL,
				)
			}
		}
	}
}

var adminDeleteReservationTests = []struct {
	name                 string
	queryParams          string
	id									 string
	src									 string
	expectedStatusCode 	 int
	expectedRedirectURL  string
}{
	{
		"Deletes reservation",
		"",
		"1",
		"calendar",
		http.StatusSeeOther,
		"/admin/reservations-calendar",
	},
	{
		"Deletes reservation and navigates user back to appropriate month and year in reservations calendar",
		"?y=2021&m=12",
		"1",
		"calendar",
		http.StatusSeeOther,
		"/admin/reservations-calendar?y=2021&m=12",
	},
	{
		"Deletes reservation and redirects user back to all-reservations page",
		"",
		"1",
		"all",
		http.StatusSeeOther,
		"/admin/all-reservations",
	},
	{
		"Invalid id URL parameter",
		"",
		"invalid",
		"calendar",
		http.StatusInternalServerError,
		"",
	},
	{
		"Delete reservation failed",
		"",
		"11",
		"calendar",
		http.StatusInternalServerError,
		"",
	},
}

func TestRepository_AdminDeleteReservation(t *testing.T) {
	for _, test := range adminDeleteReservationTests {
		// Create GET request to `/admin/delete-reservation/{src}/{id}`
		// and store context on it which includes the `X-Session` header
		// in order to read to/from the `Session object`
		req, err := http.NewRequest("GET", fmt.Sprintf("/admin/delete-reservation/%s/%s%s", test.src, test.id, test.queryParams), nil)
		if err != nil {
			log.Println(err)
		}

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("src", test.src)
		rctx.URLParams.Add("id", test.id)

		ctx := getRequestContext(req)
		req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// This fakes all of the request/response lifecycle
		// Stores the response we get from the request
		responseRecorder := httptest.NewRecorder()

		// Make `/admin/delete-reservation/{src}/{id}` POST handler function able to be called directly
		// and execute it
		handler := http.HandlerFunc(Repo.AdminDeleteReservation)
		handler.ServeHTTP(responseRecorder, req)

		if responseRecorder.Code != test.expectedStatusCode {
			t.Errorf(
				"Test %s returns wrong response status code: got %d, wanted %d",
				test.name,
				responseRecorder.Code,
				test.expectedStatusCode,
			)
		}

		if test.expectedRedirectURL != "" {
			// Get redirect URL
			redirectURL, err := responseRecorder.Result().Location()
			if err != nil {
				log.Println(err)
			}

			if redirectURL.String() != test.expectedRedirectURL {
				t.Errorf(
					"Test %s redirects user to wrong URL: got %s, wanted %s",
					test.name,
					redirectURL.String(),
					test.expectedRedirectURL,
				)
			}
		}
	}
}