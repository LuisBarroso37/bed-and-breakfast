package main

import (
	"encoding/gob"
	"flag"
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

	// Close SQL connection pool and mail channel
	defer pool.SQL.Close()
	defer close(app.MailChan)

	// Run mail server to listen for email messages
	log.Println("Starting mail server...")
	listenForMail()

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
	// Register the data types that we will store in the `Session` object
	gob.Register(models.Reservation{})
	gob.Register(models.Room{})
	gob.Register(models.RoomRestriction{})
	gob.Register(models.User{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	// Instantiate and store a channel to send email messages
	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	// Get configuration from env variables
	inProduction := flag.Bool("production", true, "Aplication is in production")
	useCache := flag.Bool("cache", true, "Use template cache")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database user")
	dbPassword := flag.String("dbpassword", "", "Database password")
	dbPort := flag.String("dbport", "5432", "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database SSL settings (disable, prefer, require)")

	flag.Parse()

	// Make sure all required env variables are set
	if *dbName == "" || *dbUser == "" || *dbPassword == "" {
		fmt.Println("Missing required flags")
		os.Exit(1)
	}

	// Change this to true when in production
	app.InProduction = *inProduction

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
	connectionString := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		*dbHost, 
		*dbPort, 
		*dbName, 
		*dbUser, 
		*dbPassword,
		*dbSSL,
	)
	pool, err := driver.ConnectSQL(connectionString)
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
	app.UseCache = *useCache

	// Store app configuration in 'render' package
	render.StoreAppConfig(&app)

	// Create a repository and set it in the 'handlers' package
	repo := handlers.NewRepository(&app, pool)
	handlers.SetRepository(repo)

	// Store app configuration in 'helpers' package
	helpers.StoreAppConfig(&app)

	return pool, nil
}
