package main

import (
	"fmt"
	"testing"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/config"
	"github.com/go-chi/chi/v5"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)

	switch varType := mux.(type) {
		case *chi.Mux:
			// Test passes
		default:
			// Test fails
			t.Error(fmt.Sprintf("Type should be http.Handler but instead is %T", varType))
		} 
}