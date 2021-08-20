package main

import (
	"net/http"

	"github.com/justinas/nosurf"
)

// Adds CSRF protection to all POST requests
func CreateCsrfHandler(next http.Handler) http.Handler {
	// If CSRF check is successful, `csrfHandler` calls `next`
	csrfHandler := nosurf.New(next)
		
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path: "/",
		Secure: app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
		
	return csrfHandler
}

// Loads and saves the session on every request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}