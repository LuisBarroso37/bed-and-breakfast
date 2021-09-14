package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"

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

// Unexpected errors
func ServerError(w http.ResponseWriter, err error) {
	// Stack trace and error message
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}