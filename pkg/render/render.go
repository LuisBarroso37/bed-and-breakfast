package render

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/LuisBarroso37/bed-and-breakfast/pkg/config"
	"github.com/LuisBarroso37/bed-and-breakfast/pkg/models"
)

// Custom functions passed to the GO templates
var functions = template.FuncMap{}

var app *config.AppConfig

// Store app configuration
func StoreAppConfig(appConfig *config.AppConfig) {
	app = appConfig
}

// Get all template pages
func GetTemplatePages() (map[string]*template.Template, error) {
	// Store all template pages found
	templates := map[string]*template.Template{}

	// Get all template page file paths
	pages, err  := filepath.Glob("./templates/*.page.tmpl")
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
		matches, err := filepath.Glob("./templates/*.layout.tmpl")
		if err != nil {
			return templates, err
		}

  	// If any layout files have been found, associate them with the created template page
		if len(matches) > 0 {
			template, err = template.ParseGlob("./templates/*.layout.tmpl")
			if err != nil {
				return templates, err
			}
		}

		templates[name] = template
	}

	return templates, nil
}

func addDefaultData(templateData *models.TemplateData) *models.TemplateData {
	return templateData
}

// Renders templates using HTML template
func RenderTemplate(w http.ResponseWriter, templateName string, templateData *models.TemplateData) {
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
        log.Fatal("Could not get template from template cache")
    }

    // Convert the template into bytes so we can write the data to 'ResponseWriter'
		td := addDefaultData(templateData)
    buffer := new(bytes.Buffer)
    err := template.Execute(buffer, td) // Pass `templateData` to the buffer
    if err != nil {
        fmt.Println("Failed to execute template: ", err)
	}

    // Write buffer to 'ResponseWriter'
    _, err = buffer.WriteTo(w)
    if err != nil {
        fmt.Println("Failed to execute template: ", err)
	}
}