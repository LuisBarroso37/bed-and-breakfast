package render

import (
	"net/http"
	"testing"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
)

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData

	// Create an HTTP GET request to pass to `AddDefaultData`
	request, err := getSession()
	if err != nil {
		t.Error(err)
	}

	// Add data to Session object
	session.Put(request.Context(), "success", "123")

	templateData := addDefaultData(&td, request)
	if templateData.Success != "123" {
		t.Error("Success message of '123' not found in session")
	}
}

func TestRenderTemplate(t *testing.T) {
	pathToTemplates = "../../templates"

	// Create template cache
	templateCache, err := GetTemplatePages()
	if err != nil {
		t.Error(err)
	}

	// Store template cache in global config
	app.TemplateCache = templateCache

	// Create HTTP GET request
	request, err := getSession()
	if err != nil {
		t.Error(err)
	}

	// Set test ResponseWriter
	var tw testResponseWriter
	
	// Render existing template
	err = RenderTemplate(&tw, request, "home.page.tmpl", &models.TemplateData{})
	if err != nil {
		t.Error("Error writing template to browser")
	}

	// Render non-existing template
	err = RenderTemplate(&tw, request, "non-existent.page.tmpl", &models.TemplateData{})
	if err == nil {
		t.Error("Rendered template that does not exist")
	}
}

func TestStoreAppConfig(t *testing.T) {
	StoreAppConfig(app)
}

func TestGetTemplatePages(t *testing.T) {
	pathToTemplates = "../../templates"

	_, err := GetTemplatePages()
	if err != nil {
		t.Error(err)
	}
}

func getSession() (*http.Request, error) {
	// Create an HTTP GET request to pass to `AddDefaultData`
	request, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		return nil, err
	}

	// Add Session object to request context
	ctx := request.Context()
	ctx, _ = session.Load(ctx, request.Header.Get("X-Session"))
	request = request.WithContext(ctx)

	return request, nil
}