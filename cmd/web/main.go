package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/config"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/driver"
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
	pool, err := run()
	if err != nil {
		log.Fatal(err)
	}

	defer pool.SQL.Close()

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

func run() (*driver.DB, error) {
	// What we want to store in the session in global config
	gob.Register(models.Reservation{})
	gob.Register(models.Room{})
	gob.Register(models.RoomRestriction{})
	gob.Register(models.User{})
	gob.Register(models.Restriction{})
 
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

	// Connect to database
	log.Println("Connecting to database...")
	pool, err := driver.ConnectSQL("host=localhost port=5432 dbname=bookings user=postgres password=password")
	if err != nil {
		log.Fatal("Cannot connect to database")
	}
	log.Println("Connected to database")

	// Get all template pages
	templates, err := render.GetTemplatePages()
	if err != nil {
		log.Fatal("Cannot get template pages")
		
		return nil, err
	}

	// Store template pages in the app cache
	app.TemplateCache = templates
	app.UseCache = false
	
	// Store app configuration in 'render' package
	render.StoreAppConfig(&app)

	// Create a repository and set it in the 'handlers' package
	repo := handlers.NewRepository(&app, pool)
	handlers.SetRepository(repo)

	// Store app configuration in 'helpers' package
	helpers.StoreAppConfig(&app)

	return pool, nil
}
