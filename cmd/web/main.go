package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/config"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/handlers"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/helpers"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

// App configuration
var app config.AppConfig
var session *scs.SessionManager

func main() {
	// Setup global configuration
	err := run()
	if err != nil {
		log.Fatal(err)
	}

  // Create server
	server := &http.Server{
		Addr: portNumber,
    Handler: routes(&app),
	}

  // Run server
  fmt.Printf("Application is listening on port %s \n", portNumber)
  err = server.ListenAndServe()
  if err != nil {
      log.Fatal(err)
  }
}

func run() error {
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
	templates, err := render.GetTemplatePages()
	if err != nil {
		log.Fatal("Cannot get template pages")
		
		return err
	}

	// Store template pages in the app cache
	app.TemplateCache = templates
	app.UseCache = false
	
	// Store app configuration in 'render' package
	render.StoreAppConfig(&app)

	// Create a repository and set it in the 'handlers' package
	repo := handlers.NewRepository(&app)
	handlers.SetRepository(repo)

	// Store app configuration in 'helpers' package
	helpers.StoreAppConfig(&app)

	return nil
}
