package forms

type errors map[string][]string

// Adds an error message for a given field
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Returns first error message for a given field
func (e errors) Get(field string) string {
	errorString := e[field]

	if len(errorString) == 0 {
		return ""
	}

	return errorString[0]
}