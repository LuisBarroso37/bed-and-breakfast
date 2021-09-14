package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNoSurf(t *testing.T) {
	var testHandler testHandler
	csrfHandler := CreateCsrfHandler(&testHandler)

	// Match on the variable type
	switch varType := csrfHandler.(type) {
		case http.Handler:
			// Test passes
		default:
			// Test fails
			t.Error(fmt.Sprintf("Type should be http.Handler but instead is %T", varType))
	}
}

func TestSessionLoad(t *testing.T) {
	var testHandler testHandler
	httpHandler := SessionLoad(&testHandler)

	// Match on the variable type
	switch varType := httpHandler.(type) {
		case http.Handler:
			// Test passes
		default:
			// Test fails
			t.Error(fmt.Sprintf("Type should be http.Handler but instead is %T", varType))
	}
}