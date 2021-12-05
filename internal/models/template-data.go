package models

import "github.com/LuisBarroso37/bed-and-breakfast/internal/forms"

// Holds data sent from handlers to templates
type TemplateData struct {
	StringMap map[string]string
	IntMap map[string]int
	FloatMap map[string]float32
	Data map[string]interface{}
	CsrfToken string
	Success string
	Warning string
	Error string
	Form *forms.Form
	IsAuthenticated bool
}