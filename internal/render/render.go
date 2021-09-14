package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/config"
	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
	"github.com/justinas/nosurf"
)

// Custom functions passed to the GO templates
var functions = template.FuncMap{}

var app *config.AppConfig
var pathToTemplates = "./templates"

// Store app configuration
func StoreAppConfig(appConfig *config.AppConfig) {
	app = appConfig
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

func addDefaultData(templateData *models.TemplateData, r *http.Request) *models.TemplateData {
	templateData.CsrfToken = nosurf.Token(r)
	templateData.Success = app.Session.PopString(r.Context(), "success")
	templateData.Warning = app.Session.PopString(r.Context(), "warning")
	templateData.Error = app.Session.PopString(r.Context(), "error")
	
	return templateData
}

// Renders templates using HTML template
func RenderTemplate(w http.ResponseWriter, r *http.Request, templateName string, templateData *models.TemplateData) error {
		var templates map[string]*template.Template

		// Only use cache in production mode
		// In development mode we want to see our changes in real time
		if app.UseCache {
			templates = app.TemplateCache
		} else {
			templates, _ = GetTemplatePages()
		}

    // Check if given template name exists in the 'templates' map
    template, ok := templates[templateName]
    if !ok {
				return errors.New("can't get template from template cache")
    }

    // Convert the template into bytes so we can write the data to 'ResponseWriter'
		td := addDefaultData(templateData, r)
    buffer := new(bytes.Buffer)
    err := template.Execute(buffer, td) // Pass `templateData` to the buffer
    if err != nil {
			log.Fatal(err)
		}

    // Write buffer to 'ResponseWriter'
    _, err = buffer.WriteTo(w)
    if err != nil {
        fmt.Println("Failed to execute template: ", err)
				return err
	}

	return nil
}