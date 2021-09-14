package render

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/config"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
	"github.com/alexedwards/scs/v2"
)

var session *scs.SessionManager
var testApp config.AppConfig

func TestMain(m *testing.M) {
	// What we want to store in the session in global config
	gob.Register(models.Reservation{})

	// Change this to true when in production
	testApp.InProduction = false

	// Setup info and error loggers
	testApp.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	testApp.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	
	// Session management
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false

	// Set session in global config
	testApp.Session = session

	// Set main variable `app` in render.go to reference the one created in this test setup
	app = &testApp

	os.Exit(m.Run())
}

// Mocked ResponseWriter interface
type testResponseWriter struct {}

func (tw *testResponseWriter) Header() http.Header {
	var header http.Header

	return header
}

func (tw *testResponseWriter) WriteHeader(i int) {}

func (tw *testResponseWriter) Write(b []byte) (int, error) {
	length := len(b)

	return length, nil
}