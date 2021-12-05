package config

import (
	"html/template"
	"log"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
	"github.com/alexedwards/scs/v2"
)

// Holds the application configuration
type AppConfig struct {
	UseCache 			bool
	TemplateCache map[string]*template.Template
	InProduction 	bool
	InfoLog 			*log.Logger
	ErrorLog 			*log.Logger
	Session 			*scs.SessionManager
	MailChan 			chan models.MailData
}