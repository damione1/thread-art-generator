package components

import (
	"html/template"
	"strings"

	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
)

// FormErrorData represents data for rendering form errors
type FormErrorData struct {
	FieldErrors map[string][]string
	GlobalError string
	Success     bool
}

// NewFormErrorData creates FormErrorData from a StandardError
func NewFormErrorData(err *pbErrors.StandardError) *FormErrorData {
	if err == nil {
		return &FormErrorData{
			FieldErrors: make(map[string][]string),
			Success:     true,
		}
	}

	return &FormErrorData{
		FieldErrors: err.Fields,
		GlobalError: err.GlobalError,
		Success:     false,
	}
}

// HasFieldError checks if a specific field has errors
func (f *FormErrorData) HasFieldError(field string) bool {
	errors, exists := f.FieldErrors[field]
	return exists && len(errors) > 0
}

// GetFieldError returns the first error message for a field
func (f *FormErrorData) GetFieldError(field string) string {
	if errors, exists := f.FieldErrors[field]; exists && len(errors) > 0 {
		return errors[0]
	}
	return ""
}

// GetFieldErrors returns all error messages for a field
func (f *FormErrorData) GetFieldErrors(field string) []string {
	if errors, exists := f.FieldErrors[field]; exists {
		return errors
	}
	return []string{}
}

// GetFieldErrorsAsString returns all error messages for a field as a single string
func (f *FormErrorData) GetFieldErrorsAsString(field string) string {
	errors := f.GetFieldErrors(field)
	return strings.Join(errors, ", ")
}

// HasGlobalError checks if there's a global error
func (f *FormErrorData) HasGlobalError() bool {
	return f.GlobalError != ""
}

// GetFieldClasses returns CSS classes for a form field based on error state
func (f *FormErrorData) GetFieldClasses(field string, baseClasses string) template.HTMLAttr {
	classes := baseClasses
	if f.HasFieldError(field) {
		classes += " border-red-500 focus:border-red-500 focus:ring-red-500"
	} else {
		classes += " border-gray-300 focus:border-blue-500 focus:ring-blue-500"
	}
	return template.HTMLAttr(classes)
}
