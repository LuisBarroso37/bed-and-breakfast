package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/driver"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
)

func getRequestContext(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}

	return ctx
}

// This contains only route handlers which do not require a `Session` object
var tests = []struct {
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
}

func TestHandlersThatDoNotRequireSession(t *testing.T) {
	// Setup test
	routes := getRoutes()

	// Create test server
	testServer := httptest.NewTLSServer(routes)
	defer testServer.Close()

	for _, test := range tests {
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

func TestNewRepository(t *testing.T) {
	var db driver.DB
	testRepo := NewRepository(&app, &db)

	// Check if variable returned by NewRepository() is of type *Repository
	testRepoType := reflect.TypeOf(testRepo).String()
	if testRepoType != "*handlers.Repository" {
		t.Errorf("Did not get correct type from NewRepository: got %s, wanted *Repository", testRepoType)
	}
}

func TestRepository_MakeReservation_Success(t *testing.T) {
	// Create dummy reservation
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID: 1,
			RoomName: "General's Quarters",
		},
	}

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

	// Store dummy reservation in the `Session` object
	session.Put(ctx, "reservation", reservation)

	// Make `make-reservation` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.MakeReservation)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 200
	if responseRecorder.Code != http.StatusOK {
		t.Errorf(
			"MakeReservation handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusOK,
		)
	}
}

func TestRepository_MakeReservation_ReservationNotInSession(t *testing.T) {
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

	// Make `make-reservation` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.MakeReservation)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"MakeReservation handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}
}

func TestRepository_MakeReservation_RoomDoesNotExist(t *testing.T) {
	// Create dummy reservation with room id which does not exist
	reservation := models.Reservation{
		RoomID: 3,
		Room: models.Room{
			ID: 3,
			RoomName: "General's Quarters",
		},
	}

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

	// Store dummy reservation in the `Session` object
	session.Put(ctx, "reservation", reservation)

	// Make `make-reservation` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.MakeReservation)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"MakeReservation handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}
}

func TestRepository_PostMakeReservation_Success(t *testing.T) {
	// Build request body
	body := url.Values{}
	body.Add("start_date", "2050-01-01")
	body.Add("end_date", "2050-01-02")
	body.Add("first_name", "John")
	body.Add("last_name", "Smith")
	body.Add("email", "john@smith.com")
	body.Add("phone", "123456789")
	body.Add("room_id", "1")

	// Create POST request to `/make-reservation` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/make-reservation", strings.NewReader(body.Encode()))
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Make `make-reservation` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 303
	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf(
			"PostMakeReservation handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusSeeOther,
		)
	}
}

func TestRepository_PostMakeReservation_UnableToParseForm(t *testing.T) {
	// Create POST request to `/make-reservation` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/make-reservation", nil)
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Make `make-reservation` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"PostMakeReservation handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}
}

func TestRepository_PostMakeReservation_InvalidDates(t *testing.T) {
	// ----------- Invalid start date ------------

	// Build request body
	body := url.Values{}
	body.Add("start_date", "invalid")
	body.Add("end_date", "2050-01-02")
	body.Add("first_name", "John")
	body.Add("last_name", "Smith")
	body.Add("email", "john@smith.com")
	body.Add("phone", "123456789")
	body.Add("room_id", "1")

	// Create POST request to `/make-reservation` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/make-reservation", strings.NewReader(body.Encode()))
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Make `make-reservation` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"PostMakeReservation handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}

	// ----------- Invalid end date ------------

	// Build request body
	body = url.Values{}
	body.Add("start_date", "2050-01-01")
	body.Add("end_date", "invalid")
	body.Add("first_name", "John")
	body.Add("last_name", "Smith")
	body.Add("email", "john@smith.com")
	body.Add("phone", "123456789")
	body.Add("room_id", "1")

	// Create POST request to `/make-reservation` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err = http.NewRequest("POST", "/make-reservation", strings.NewReader(body.Encode()))
	if err != nil {
		log.Println(err)
	}
	ctx = getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder = httptest.NewRecorder()

	// Make `make-reservation` handler function able to be called directly
	// and execute it
	handler = http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"PostMakeReservation handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}
}

func TestRepository_PostMakeReservation_InvalidRoomID(t *testing.T) {
	// Build request body
	body := url.Values{}
	body.Add("start_date", "2050-01-01")
	body.Add("end_date", "2050-01-02")
	body.Add("first_name", "John")
	body.Add("last_name", "Smith")
	body.Add("email", "john@smith.com")
	body.Add("phone", "123456789")
	body.Add("room_id", "invalid")

	// Create POST request to `/make-reservation` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/make-reservation", strings.NewReader(body.Encode()))
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Make `make-reservation` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"PostMakeReservation handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}
}

func TestRepository_PostMakeReservation_InvalidFormData(t *testing.T) {
	// Build request body
	body := url.Values{}
	body.Add("start_date", "2050-01-01")
	body.Add("end_date", "2050-01-02")
	body.Add("first_name", "L")
	body.Add("last_name", "Xiang")
	body.Add("email", "l@xiang.com")
	body.Add("phone", "123456789")
	body.Add("room_id", "1")

	// Create POST request to `/make-reservation` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/make-reservation", strings.NewReader(body.Encode()))
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Make `make-reservation` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 303
	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf(
			"PostMakeReservation handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusSeeOther,
		)
	}
}

func TestRepository_PostMakeReservation_FailureToInsertReservationInDB(t *testing.T) {
	// Build request body
	body := url.Values{}
	body.Add("start_date", "2050-01-01")
	body.Add("end_date", "2050-01-02")
	body.Add("first_name", "John")
	body.Add("last_name", "Smith")
	body.Add("email", "john@smith.com")
	body.Add("phone", "123456789")
	body.Add("room_id", "2")

	// Create POST request to `/make-reservation` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/make-reservation", strings.NewReader(body.Encode()))
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Make `make-reservation` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"PostMakeReservation handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}
}

func TestRepository_PostMakeReservation_FailureToInsertRoomRestrictionInDB(t *testing.T) {
	// Build request body
	body := url.Values{}
	body.Add("start_date", "2050-01-01")
	body.Add("end_date", "2050-01-02")
	body.Add("first_name", "John")
	body.Add("last_name", "Smith")
	body.Add("email", "john@smith.com")
	body.Add("phone", "123456789")
	body.Add("room_id", "1000")

	// Create POST request to `/make-reservation` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/make-reservation", strings.NewReader(body.Encode()))
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Make `make-reservation` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.PostMakeReservation)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"PostMakeReservation handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}
}

func TestRepository_PostSearchAvailability_Success(t *testing.T) {
	// Build request body with valid start date -- before 2049-12-31
	body := url.Values{}
	body.Add("start_date", "2049-01-01")
	body.Add("end_date", "2049-01-02")

	// Create POST request to `/search-availability` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/search-availability", strings.NewReader(body.Encode()))
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

	// Throw error if status code received in the response is not a 200
	if responseRecorder.Code != http.StatusOK {
		t.Errorf(
			"PostSearchAvailability handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusOK,
		)
	}
}

func TestRepository_PostSearchAvailability_UnableToParseForm(t *testing.T) {
	// Create POST request to `/search-availability` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/search-availability", nil)
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

	// Throw error if status code received in the response is not a 200
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"PostSearchAvailability handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}
}

func TestRepository_PostSearchAvailability_InvalidDates(t *testing.T) {
	// ----------- Invalid start date ------------

	// Build request body
	body := url.Values{}
	body.Add("start_date", "invalid")
	body.Add("end_date", "2049-01-02")

	// Create POST request to `/search-availability` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/search-availability", strings.NewReader(body.Encode()))
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

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"PostMakeReservation handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}

	// ----------- Invalid end date ------------

	// Build request body
	body = url.Values{}
	body.Add("start_date", "2049-01-02")
	body.Add("end_date", "invalid")

	// Create POST request to `/search-availability` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err = http.NewRequest("POST", "/search-availability", strings.NewReader(body.Encode()))
	if err != nil {
		log.Println(err)
	}
	ctx = getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder = httptest.NewRecorder()

	// Make `search-availability` handler function able to be called directly
	// and execute it
	handler = http.HandlerFunc(Repo.PostSearchAvailability)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"PostMakeReservation handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}
}

func TestRepository_PostSearchAvailability_DatabaseQueryFails(t *testing.T) {
	// Build request body with invalid start date of 2000-01-01, which will cause
	// our test db repository to return an error
	body := url.Values{}
	body.Add("start_date", "2000-01-01")
	body.Add("end_date", "2000-01-02")

	// Create POST request to `/search-availability` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/search-availability", strings.NewReader(body.Encode()))
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

	// Throw error if status code received in the response is not a 200
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"PostSearchAvailability handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}
}

func TestRepository_PostSearchAvailability_RoomsNotAvailable(t *testing.T) {
	// Build request body with start date after of 2049-12-31, which will cause
	// our test db repository to return no available rooms
	body := url.Values{}
	body.Add("start_date", "2050-01-01")
	body.Add("end_date", "2050-01-02")

	// Create POST request to `/search-availability` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/search-availability", strings.NewReader(body.Encode()))
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

	// Throw error if status code received in the response is not a 200
	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf(
			"PostSearchAvailability handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusSeeOther,
		)
	}
}

func TestRepository_SearchAvailabilityJson_Success(t *testing.T) {
	// Build request body with valid dates -- before 2049-12-31
	body := url.Values{}
	body.Add("start_date", "2049-01-01")
	body.Add("end_date", "2049-01-02")
	body.Add("room_id", "1")

	// Create POST request to `/search-availability-json` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/search-availability-json", strings.NewReader(body.Encode()))
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Make `search-availability-json` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.SearchAvailabilityJson)
	handler.ServeHTTP(responseRecorder, req)

	var jsonRes jsonResponse
	err = json.Unmarshal(responseRecorder.Body.Bytes(), &jsonRes)
	if err != nil {
		t.Error("Failed to parse json response")
	}

	// Throw error if `OK` property is set to false in the json response
	if !jsonRes.OK {
		t.Error("SearchAvailabilityJson gives no availability when it is expected")
	}
}

func TestRepository_SearchAvailabilityJson_RoomsNotAvailable(t *testing.T) {
	// Build request body with invalid dates -- after 2049-12-31
	body := url.Values{}
	body.Add("start_date", "2050-01-01")
	body.Add("end_date", "2050-01-02")
	body.Add("room_id", "1")

	// Create POST request to `/search-availability-json` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/search-availability-json", strings.NewReader(body.Encode()))
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Make `search-availability-json` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.SearchAvailabilityJson)
	handler.ServeHTTP(responseRecorder, req)

	var jsonRes jsonResponse
	err = json.Unmarshal(responseRecorder.Body.Bytes(), &jsonRes)
	if err != nil {
		t.Error("Failed to parse json response")
	}

	// Throw error if `OK` property is set to true in the json response
	if jsonRes.OK {
		t.Error("SearchAvailabilityJson gives back availability when none was expected")
	}
}

func TestRepository_SearchAvailabilityJson_UnableToParseForm(t *testing.T) {
	// Create POST request to `/search-availability-json` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/search-availability-json", nil)
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Make `search-availability-json` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.SearchAvailabilityJson)
	handler.ServeHTTP(responseRecorder, req)

	var jsonRes jsonResponse
	err = json.Unmarshal(responseRecorder.Body.Bytes(), &jsonRes)
	if err != nil {
		t.Error("Failed to parse json response")
	}

	// Throw error if `OK` property is set to true in the json response
	if jsonRes.OK || jsonRes.Message != "Internal Server Error" {
		t.Error("SearchAvailabilityJson does not throw error when one is expected")
	}
}

func TestRepository_SearchAvailabilityJson_InvalidDates(t *testing.T) {
	// ----------- Invalid start date ------------

	// Build request body
	body := url.Values{}
	body.Add("start_date", "invalid")
	body.Add("end_date", "2050-01-02")
	body.Add("room_id", "1")

	// Create POST request to `/search-availability-json` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/search-availability-json", strings.NewReader(body.Encode()))
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Make `search-availability-json` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.SearchAvailabilityJson)
	handler.ServeHTTP(responseRecorder, req)

	var jsonRes jsonResponse
	err = json.Unmarshal(responseRecorder.Body.Bytes(), &jsonRes)
	if err != nil {
		t.Error("Failed to parse json response")
	}

	// Throw error if `OK` property is set to true in the json response
	if jsonRes.OK || jsonRes.Message != "Internal Server Error"  {
		t.Error("SearchAvailabilityJson does not throw error when one is expected")
	}

	// ----------- Invalid end date ------------

	// Build request body
	body = url.Values{}
	body.Add("start_date", "2050-01-01")
	body.Add("end_date", "invalid")
	body.Add("room_id", "1")

	// Create POST request to `/search-availability-json` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err = http.NewRequest("POST", "/search-availability-json", strings.NewReader(body.Encode()))
	if err != nil {
		log.Println(err)
	}
	ctx = getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder = httptest.NewRecorder()

	// Make `search-availability-json` handler function able to be called directly
	// and execute it
	handler = http.HandlerFunc(Repo.SearchAvailabilityJson)
	handler.ServeHTTP(responseRecorder, req)

	err = json.Unmarshal(responseRecorder.Body.Bytes(), &jsonRes)
	if err != nil {
		t.Error("Failed to parse json response")
	}

	// Throw error if `OK` property is set to true in the json response
	if jsonRes.OK || jsonRes.Message != "Internal Server Error"  {
		t.Error("SearchAvailabilityJson does not throw error when one is expected")
	}
}

func TestRepository_SearchAvailabilityJson_InvalidRoomID(t *testing.T) {
	// Build request body with valid dates -- before 2049-12-31 -- and invalid room id
	body := url.Values{}
	body.Add("start_date", "2049-01-01")
	body.Add("end_date", "2049-01-02")
	body.Add("room_id", "invalid")

	// Create POST request to `/search-availability-json` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/search-availability-json", strings.NewReader(body.Encode()))
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Make `search-availability-json` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.SearchAvailabilityJson)
	handler.ServeHTTP(responseRecorder, req)

	var jsonRes jsonResponse
	err = json.Unmarshal(responseRecorder.Body.Bytes(), &jsonRes)
	if err != nil {
		t.Error("Failed to parse json response")
	}

	// Throw error if `OK` property is set to true in the json response
	if jsonRes.OK || jsonRes.Message != "Internal Server Error"  {
		t.Error("SearchAvailabilityJson does not throw error when one is expected")
	}
}

func TestRepository_SearchAvailabilityJson_DatabaseQueryFails(t *testing.T) {
	// Build request body with invalid start date of 2000-01-01, which will cause
	// our test db repository to return an error
	body := url.Values{}
	body.Add("start_date", "2000-01-01")
	body.Add("end_date", "2000-01-02")
	body.Add("room_id", "1")

	// Create POST request to `/search-availability-json` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("POST", "/search-availability-json", strings.NewReader(body.Encode()))
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Make `search-availability-json` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.SearchAvailabilityJson)
	handler.ServeHTTP(responseRecorder, req)

	var jsonRes jsonResponse
	err = json.Unmarshal(responseRecorder.Body.Bytes(), &jsonRes)
	if err != nil {
		t.Error("Failed to parse json response")
	}

	// Throw error if `OK` property is set to true in the json response
	if jsonRes.OK || jsonRes.Message != "Error connecting to database" {
		t.Error("SearchAvailabilityJson does not throw error when one is expected")
	}
}

func TestRepository_ReservationSummary_Success(t *testing.T) {
	// Create dummy reservation
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID: 1,
			RoomName: "General's Quarters",
		},
	}

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

	// Store dummy reservation in the `Session` object
	session.Put(ctx, "reservation", reservation)

	// Make `reservation-summary` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.ReservationSummary)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 200
	if responseRecorder.Code != http.StatusOK {
		t.Errorf(
			"ReservationSummary handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusOK,
		)
	}
}

func TestRepository_ReservationSummary_NoReservationInSession(t *testing.T) {
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

	// Make `reservation-summary` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.ReservationSummary)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 200
	if responseRecorder.Code == http.StatusOK {
		t.Errorf(
			"ReservationSummary handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusOK,
		)
	}
}

func TestRepository_ChooseRoom_Success(t *testing.T) {
	// Create dummy reservation
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID: 1,
			RoomName: "General's Quarters",
		},
	}

	// Create http request to `/choose-room/1` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("GET", "/choose-room/1", nil)
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)

	// Set the RequestURI on the request so that we can grab the ID
	// from the URL
	req.RequestURI = "/choose-room/1"

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Store dummy reservation in the `Session` object
	session.Put(ctx, "reservation", reservation)

	// Make `choose-room` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 303
	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf(
			"ChooseRoom handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusSeeOther,
		)
	}
}

func TestRepository_ChooseRoom_MissingOrInvalidUrlParameter(t *testing.T) {
	// ----------- Missing URL parameter ---------------

	// Create dummy reservation
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID: 1,
			RoomName: "General's Quarters",
		},
	}

	// Create http request to `/choose-room/1` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("GET", "/choose-room", nil)
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)
	
	// Set the RequestURI on the request so that we can grab the ID
	// from the URL
	req.RequestURI = "/choose-room"

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Store dummy reservation in the `Session` object
	session.Put(ctx, "reservation", reservation)

	// Make `choose-room` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"ChooseRoom handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}

	// ----------- Invalid URL parameter ---------------

	// Create http request to `/choose-room/1` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err = http.NewRequest("GET", "/choose-room/invalid", nil)
	if err != nil {
		log.Println(err)
	}
	ctx = getRequestContext(req)
	req = req.WithContext(ctx)
	
	// Set the RequestURI on the request so that we can grab the ID
	// from the URL
	req.RequestURI = "/choose-room/invalid"

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder = httptest.NewRecorder()

	// Make `choose-room` handler function able to be called directly
	// and execute it
	handler = http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"ChooseRoom handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}
}

func TestRepository_ChooseRoom_NoReservationInSession(t *testing.T) {
	// Create http request to `/choose-room/1` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("GET", "/choose-room/1", nil)
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)

	// Set the RequestURI on the request so that we can grab the ID
	// from the URL
	req.RequestURI = "/choose-room/1"

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Make `choose-room` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"ChooseRoom handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}
}

func TestRepository_BookRoom_Success(t *testing.T) {
	// Create dummy reservation
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID: 1,
			RoomName: "General's Quarters",
		},
	}

	// Create http request to `/book-room` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("GET", "/book-room?start_date=2049-01-01&end_date=2049-01-02&id=1", nil)
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Store dummy reservation in the `Session` object
	session.Put(ctx, "reservation", reservation)

	// Make `book-room` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 303
	if responseRecorder.Code != http.StatusSeeOther {
		t.Errorf(
			"ChooseRoom handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusSeeOther,
		)
	}
}

func TestRepository_BookRoom_MissingOrInvalidIdUrlParameter(t *testing.T) {
	// ----------- Missing URL parameter ---------------

	// Create dummy reservation
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID: 1,
			RoomName: "General's Quarters",
		},
	}

	// Create http request to `/book-room` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("GET", "/book-room?start_date=2049-01-01&end_date=2049-01-02", nil)
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Store dummy reservation in the `Session` object
	session.Put(ctx, "reservation", reservation)

	// Make `book-room` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"ChooseRoom handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}

	// ----------- Invalid URL parameter ---------------

	// Create http request to `/book-room` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err = http.NewRequest("GET", "/book-room?start_date=2049-01-01&end_date=2049-01-02&id=invalid", nil)
	if err != nil {
		log.Println(err)
	}
	ctx = getRequestContext(req)
	req = req.WithContext(ctx)

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder = httptest.NewRecorder()

	// Make `book-room` handler function able to be called directly
	// and execute it
	handler = http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"ChooseRoom handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}
}

func TestRepository_BookRoom_MissingOrInvalidDatesUrlParameter(t *testing.T) {
	// ----------- Missing URL parameter - start_date ---------------

	// Create dummy reservation
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID: 1,
			RoomName: "General's Quarters",
		},
	}

	// Create http request to `/book-room` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("GET", "/book-room?end_date=2049-01-02&id=1", nil)
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Store dummy reservation in the `Session` object
	session.Put(ctx, "reservation", reservation)

	// Make `book-room` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"ChooseRoom handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}

	// ----------- Missing URL parameter - end_date ---------------

	// Create http request to `/book-room` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err = http.NewRequest("GET", "/book-room?start_date=2049-01-01&id=1", nil)
	if err != nil {
		log.Println(err)
	}
	ctx = getRequestContext(req)
	req = req.WithContext(ctx)

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder = httptest.NewRecorder()

	// Store dummy reservation in the `Session` object
	session.Put(ctx, "reservation", reservation)

	// Make `book-room` handler function able to be called directly
	// and execute it
	handler = http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"ChooseRoom handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}

	// ----------- Invalid URL parameter - start_date ---------------

	// Create http request to `/book-room` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err = http.NewRequest("GET", "/book-room?start_date=invalid&end_date=2049-01-02&id=1", nil)
	if err != nil {
		log.Println(err)
	}
	ctx = getRequestContext(req)
	req = req.WithContext(ctx)

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder = httptest.NewRecorder()

	// Store dummy reservation in the `Session` object
	session.Put(ctx, "reservation", reservation)

	// Make `book-room` handler function able to be called directly
	// and execute it
	handler = http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"ChooseRoom handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}

	// ----------- Invalid URL parameter - start_date ---------------

	// Create http request to `/book-room` 
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err = http.NewRequest("GET", "/book-room?start_date=2049-01-01&end_date=invalid&id=1", nil)
	if err != nil {
		log.Println(err)
	}
	ctx = getRequestContext(req)
	req = req.WithContext(ctx)

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder = httptest.NewRecorder()

	// Store dummy reservation in the `Session` object
	session.Put(ctx, "reservation", reservation)

	// Make `book-room` handler function able to be called directly
	// and execute it
	handler = http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"ChooseRoom handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusTemporaryRedirect,
		)
	}
}

func TestRepository_BookRoom_DatabaseQueryFails(t *testing.T) {
	// Create dummy reservation
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID: 1,
			RoomName: "General's Quarters",
		},
	}

	// Create http request to `/book-room` with invalid query parameter `id` -- bigger than 2
	// and store context on it which includes the `X-Session` header
	// in order to read to/from the `Session object`
	req, err := http.NewRequest("GET", "/book-room?start_date=2049-01-01&end_date=2049-01-02&id=3", nil)
	if err != nil {
		log.Println(err)
	}
	ctx := getRequestContext(req)
	req = req.WithContext(ctx)

	// This fakes all of the request/response lifecycle
	// Stores the response we get from the request
	responseRecorder := httptest.NewRecorder()

	// Store dummy reservation in the `Session` object
	session.Put(ctx, "reservation", reservation)

	// Make `book-room` handler function able to be called directly
	// and execute it
	handler := http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(responseRecorder, req)

	// Throw error if status code received in the response is not a 307
	if responseRecorder.Code != http.StatusTemporaryRedirect {
		t.Errorf(
			"ChooseRoom handler returns wrong response status code: got %d, wanted %d",
			responseRecorder.Code,
			http.StatusSeeOther,
		)
	}
}