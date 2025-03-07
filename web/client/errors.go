package client

import (
	"fmt"
	"strings"
)

// FieldError represents an error for a specific field
type FieldError struct {
	Field   string
	Message string
}

// ValidationErrors holds all validation errors
type ValidationErrors struct {
	FieldErrors  map[string]string
	GeneralError string
}

// HasErrors returns true if there are any errors
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.FieldErrors) > 0 || ve.GeneralError != ""
}

// HasFieldError returns true if the specified field has an error
func (ve *ValidationErrors) HasFieldError(field string) bool {
	_, exists := ve.FieldErrors[field]
	return exists
}

// GetFieldError returns the error message for a field
func (ve *ValidationErrors) GetFieldError(field string) string {
	if message, exists := ve.FieldErrors[field]; exists {
		return message
	}
	return ""
}

// ParseValidationError parses an error message from the API and returns structured validation errors
func ParseValidationError(errorMessage string) *ValidationErrors {
	result := &ValidationErrors{
		FieldErrors: make(map[string]string),
	}

	// If no error or not a validation error, return empty result
	if errorMessage == "" || !strings.Contains(errorMessage, "failed to validate request") {
		if errorMessage != "" {
			result.GeneralError = errorMessage
		}
		return result
	}

	// Extract the part after the validation prefix
	parts := strings.SplitN(errorMessage, "failed to validate request: ", 2)
	if len(parts) < 2 {
		result.GeneralError = errorMessage
		return result
	}

	errorBody := parts[1]

	// Extract content inside parentheses if present
	startIdx := strings.Index(errorBody, "(")
	endIdx := strings.LastIndex(errorBody, ")")

	if startIdx >= 0 && endIdx > startIdx {
		// We found content inside parentheses, parse it for field errors
		fieldViolations := errorBody[startIdx+1 : endIdx]

		// Split by semicolon to get individual field errors
		fieldErrors := strings.Split(fieldViolations, ";")

		for _, fieldError := range fieldErrors {
			fieldError = strings.TrimSpace(fieldError)
			if fieldError == "" {
				continue
			}

			// Split by colon to get field and message
			fieldParts := strings.SplitN(fieldError, ":", 2)
			if len(fieldParts) == 2 {
				field := strings.TrimSpace(fieldParts[0])
				message := strings.TrimSpace(fieldParts[1])

				// Map backend field names to frontend field names if needed
				switch field {
				case "first_name":
					field = "name"
				case "refresh_token":
					field = "refreshToken"
				}

				result.FieldErrors[field] = message
			}
		}

		// If we parsed field errors, return the result
		if len(result.FieldErrors) > 0 {
			return result
		}
	}

	// If we couldn't extract field errors but have an error body, set it as general error
	if errorBody != "" {
		result.GeneralError = errorBody
	}

	return result
}

// Format creates a user-friendly error message for display
func (ve *ValidationErrors) Format() string {
	if ve.GeneralError != "" {
		return ve.GeneralError
	}

	if len(ve.FieldErrors) == 0 {
		return ""
	}

	// Format field errors as a list
	var messages []string
	for field, message := range ve.FieldErrors {
		messages = append(messages, fmt.Sprintf("%s: %s", field, message))
	}

	return strings.Join(messages, ", ")
}
