package templates

import "github.com/Damione1/thread-art-generator/web/client"

// HasFormLevelErrors checks if there are any errors that should be displayed at the form level
func HasFormLevelErrors(errors *client.ValidationErrors) bool {
	// List of field names that are considered form-level errors
	formLevelErrorFields := []string{"user", "session", "auth"}

	for _, field := range formLevelErrorFields {
		if errors.HasFieldError(field) {
			return true
		}
	}

	return false
}
