package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/config"
)

var app *config.AppConfig

// Stores app config for this package
func StoreAppConfig(appConfig *config.AppConfig) {
	app = appConfig
}

// Errors done by the client using our API
func ClientError(w http.ResponseWriter, status int) {
	app.InfoLog.Println("Client Error with status of", status)
	http.Error(w, http.StatusText(status), status)
}

// Unexpected server errors
func ServerError(w http.ResponseWriter, err error) {
	// Stack trace and error message
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// Parse strings into dates
func ParseDates(w http.ResponseWriter, sd string, ed string) (time.Time, time.Time, error) {
	dateFormat := "2006-01-02"

	startDate, err := time.Parse(dateFormat, sd)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endDate, err := time.Parse(dateFormat, ed)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return startDate, endDate, nil
}

// Check if user is authenticated
func IsAuthenticated(r *http.Request) bool {
	return app.Session.Exists(r.Context(), "user_id")
}