package handlers

import (
	"net/http"

	"github.com/LuisBarroso37/bed-and-breakfast/pkg/config"
	"github.com/LuisBarroso37/bed-and-breakfast/pkg/models"
	"github.com/LuisBarroso37/bed-and-breakfast/pkg/render"
)

type Repository struct {
	App *config.AppConfig
}

// Reposiotry used by the handlers
var Repo *Repository

// Create a new repository
func NewRepository(config *config.AppConfig) *Repository {
	return &Repository{ App: config }
}

// Sets the reposiotry for the handlers
func SetRepository(repo *Repository) {
	Repo = repo
}

// Home is the home page handler
func (repo *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	repo.App.Session.Put(r.Context(), "remote_ip", remoteIP)

	render.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})
}

// About is the about page handler
func (repo *Repository) About(w http.ResponseWriter, r *http.Request) {
	remoteIP := repo.App.Session.GetString(r.Context(), "remote_ip")
	stringMap := map[string]string{ "test": "Hello again!"}
	stringMap["remote_ip"] = remoteIP

	render.RenderTemplate(w, "about.page.tmpl", &models.TemplateData{ StringMap: stringMap })
}