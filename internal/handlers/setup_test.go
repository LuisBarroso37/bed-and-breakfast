package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/config"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/helpers"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/render"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
)

var app config.AppConfig
var session *scs.SessionManager
var pathToTemplates = "../../templates"

// Custom functions passed to the GO templates
var functions = template.FuncMap{}

func TestMain(m *testing.M) {
	// What we want to store in the session in global config
	gob.Register(models.Reservation{})

	// Change this to true when in production
	app.InProduction = false

	// Setup info and error loggers
	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Session management
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	// Set session in global config
	app.Session = session

	// Get all template pages
	templates, err := GetTemplatePages()
	if err != nil {
		log.Fatal("Cannot get template pages")
	}

	// Store template pages in the app cache
	app.TemplateCache = templates
	app.UseCache = true

	// Store app configuration in 'render' package
	render.StoreAppConfig(&app)

	// Create a repository and set it in the 'handlers' package
	repo := NewTestRepository(&app)
	SetRepository(repo)

	// Store app configuration in 'helpers' package
	helpers.StoreAppConfig(&app)

	os.Exit(m.Run())
}

func getRoutes() http.Handler {
	// Create router
	mux := chi.NewRouter()
	
	// Set middlewares
	mux.Use(middleware.Recoverer)
	//mux.Use(CreateCsrfHandler)
	mux.Use(SessionLoad)
	
	// Routes
	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/generals-quarters", Repo.Generals)
	mux.Get("/majors-suite", Repo.Majors)
	mux.Get("/contact", Repo.Contact)
	
	mux.Get("/search-availability", Repo.SearchAvailability)
	mux.Post("/search-availability", Repo.PostSearchAvailability)
	mux.Post("/search-availability-json", Repo.SearchAvailabilityJson)
	
	mux.Get("/make-reservation", Repo.MakeReservation)
	mux.Post("/make-reservation", Repo.PostMakeReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)
	
	// Serve static files
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	
	return mux
}

// Adds CSRF protection to all POST requests
func CreateCsrfHandler(next http.Handler) http.Handler {
	// If CSRF check is successful, `csrfHandler` calls `next`
	csrfHandler := nosurf.New(next)
		
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path: "/",
		Secure: app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
		
	return csrfHandler
}

// Loads and saves the session on every request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// Get all template pages
func GetTemplatePages() (map[string]*template.Template, error) {
	// Store all template pages found
	templates := map[string]*template.Template{}

	// Get all template page file paths
	pages, err  := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return templates, err
	}

	for _, page := range pages {
    // Get file name from file path
		name := filepath.Base(page)

    // Create template
		template, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return templates, err
		}

    // Find layout files
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return templates, err
		}

  	// If any layout files have been found, associate them with the created template page
		if len(matches) > 0 {
			template, err = template.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {
				return templates, err
			}
		}

		templates[name] = template
	}

	return templates, nil
}