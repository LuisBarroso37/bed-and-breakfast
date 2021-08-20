package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/LuisBarroso37/bed-and-breakfast/pkg/config"
	"github.com/LuisBarroso37/bed-and-breakfast/pkg/handlers"
	"github.com/LuisBarroso37/bed-and-breakfast/pkg/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

// App configuration
var app config.AppConfig
var session *scs.SessionManager

func main() {
	// Change this to true when in production
	app.InProduction = false

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
	}

	// Store template pages in the app cache
	app.TemplateCache = templates
	app.UseCache = false
	
	// Store app configuration in 'render' package
	render.StoreAppConfig(&app)

	// Create a repository and set it in the 'handlers' package
	repo := handlers.NewRepository(&app)
	handlers.SetRepository(repo)

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
