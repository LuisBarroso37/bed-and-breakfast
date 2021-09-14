package forms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

// Creates a custom form struct used for server side validation of forms
type Form struct {
	url.Values
	Errors errors
}

// Initializes a Form struct
// url.Values are the form values sent in a POST request
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Checks if form field is in POST request
func (form *Form) Has(field string) bool {
	existingField := form.Get(field)

	return existingField != ""
}

// Returns true if there are no errors, otherwise it returns false
func (form *Form) IsValid() bool {
	return len(form.Errors) == 0
}

// Checks if all require form fields are present
func (form *Form) RequiredFields(fields ...string) {
	// Loop through all received fields and check that they exist in the Form struct
	// Add errors to Form struct if they occur
	for _, field := range fields {
		value := form.Get(field)

		// Remove extra whitespace if it exists
		if strings.TrimSpace(value) == "" {
			form.Errors.Add(field, "This field cannot be empty")
		}
	}
}

// Checks if field (string) has the required minimum length
func (form *Form) MinLength(field string, length int) bool {
	// Fetch field from the request's form data
	value := form.Get(field)

	if len(value) < length {
		form.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}

	return true
}


// Checks if email is valid
func (form *Form) IsEmail(field string) {
	if !govalidator.IsEmail(form.Get(field)) {
		form.Errors.Add(field, "Invalid email address")
	}
}