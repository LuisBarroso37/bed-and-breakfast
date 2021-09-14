package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	// POST request with empty data object 
	// Nothing to validate so `IsValid` should pass
	r := httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	if !form.IsValid() {
		t.Error("Got invalid when it should have been valid")
	}
}

func TestForm_Required(t *testing.T) {
	// Create form data object
	postedData := url.Values{}

	// Check that required fields qre missing
	form := New(postedData)
	form.RequiredFields("a", "b", "c")

	if form.IsValid() {
		t.Error("Form is valid when required fields are missing")
	}

	// Add required fields to form data object
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	// Create new form with new form data
	form = New(postedData)
	form.RequiredFields("a", "b", "c")
	if !form.IsValid() {
		t.Error("Form is invalid when all required fields are present")
	}
}

func TestForm_MinLength(t *testing.T) {
	// Create form data object
	postedData := url.Values{}
	
	// Check that minimum length is not complied with if field does not exist
	form := New(postedData)
	form.MinLength("a", 3)

	if form.IsValid() {
		t.Error("Got valid when it should have been invalid - Field does not exist")
	}

	// Test `Get` method of `errors.go` when there is an error
	isError := form.Errors.Get("a")
	if isError == "" {
		t.Error("Should have an error but did not get one")
	}

	// Add required fields to form data object
	postedData.Add("a", "a")

	// Check that minimum length is not complied with
	form = New(postedData)
	form.MinLength("a", 3)

	if form.IsValid() {
		t.Error("Got valid when it should have been invalid - field does not have minimum length")
	}

	postedData.Set("a", "aaa")

	// Check that minimum length is complied with
	form = New(postedData)
	form.MinLength("a", 3)
	
	if !form.IsValid() {
		t.Error("Got invalid when it should have been valid - field has minimum length")
	}

	// Test `Get` method of `errors.go` when there is no error
	isError = form.Errors.Get("a")
	if isError != "" {
		t.Error("Should not have an error but did got one")
	}
}

func TestForm_IsEmail(t *testing.T) {
	// Create form data object
	postedData := url.Values{}
	
	// Check that email is invalid if field does not exist
	form := New(postedData)
	form.IsEmail("a")

	if form.IsValid() {
		t.Error("Got valid when it should have been invalid - Field does not exist")
	}

	// Add required fields to form data object
	postedData.Add("email", "a@")

	// Check that email is invalid
	form = New(postedData)
	form.IsEmail("email")

	if form.IsValid() {
		t.Error("Got valid when it should have been invalid - email is invalid")
	}

	// Update required fields of form data object
	postedData.Set("email", "a@a.com")

	// Check that email is valid
	form = New(postedData)
	form.IsEmail("email")

	if !form.IsValid() {
		t.Error("Got invalid when it should have been valid - email is valid")
	}
}

func TestForm_Has(t *testing.T) {
	// Create form data object
	postedData := url.Values{}
	
	// Check that field does not exist
	form := New(postedData)
	
	if form.Has("a") {
		t.Error("Got valid when it should have been invalid - Field does not exist")
	}

	// Add required fields to form data object
	postedData.Add("a", "a")

	// Check that field exists
	form = New(postedData)

	if !form.Has("a") {
		t.Error("Got invalid when it should have been valid - field exists")
	}
}