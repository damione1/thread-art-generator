package components

import (
	"html/template"
	"strings"

	"github.com/Damione1/thread-art-generator/core/errors"
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

// ErrorHelpers provides template helper functions for error handling
type ErrorHelpers struct{}

// NewErrorHelpers creates a new ErrorHelpers instance
func NewErrorHelpers() *ErrorHelpers {
	return &ErrorHelpers{}
}

// FieldErrorHTML generates HTML for field errors
func (h *ErrorHelpers) FieldErrorHTML(errorData *FormErrorData, field string) template.HTML {
	if !errorData.HasFieldError(field) {
		return ""
	}

	errors := errorData.GetFieldErrors(field)
	html := `<div class="mt-1">`
	for _, err := range errors {
		html += `<p class="text-sm text-red-600">` + template.HTMLEscapeString(err) + `</p>`
	}
	html += `</div>`
	
	return template.HTML(html)
}

// GlobalErrorHTML generates HTML for global errors (toasts)
func (h *ErrorHelpers) GlobalErrorHTML(errorData *FormErrorData) template.HTML {
	if !errorData.HasGlobalError() {
		return ""
	}

	// Toast-style error message
	html := `<div class="fixed top-4 right-4 z-50 bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded shadow-lg">
		<div class="flex">
			<div class="py-1">
				<svg class="fill-current h-6 w-6 text-red-500 mr-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20">
					<path d="M2.93 17.07A10 10 0 1 1 17.07 2.93 10 10 0 0 1 2.93 17.07zm12.73-1.41A8 8 0 1 0 4.34 4.34a8 8 0 0 0 11.32 11.32zM9 11V9h2v6H9v-4zm0-6h2v2H9V5z"/>
				</svg>
			</div>
			<div>
				<p class="font-bold">Error</p>
				<p class="text-sm">` + template.HTMLEscapeString(errorData.GlobalError) + `</p>
			</div>
			<div class="ml-auto">
				<button type="button" class="text-red-700 hover:text-red-900" onclick="this.parentElement.parentElement.parentElement.remove()">
					<svg class="fill-current h-6 w-6" role="button" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20">
						<title>Close</title>
						<path d="M14.348 14.849a1.2 1.2 0 0 1-1.697 0L10 11.819l-2.651 3.029a1.2 1.2 0 1 1-1.697-1.697l2.758-3.15-2.759-3.152a1.2 1.2 0 1 1 1.697-1.697L10 8.183l2.651-3.031a1.2 1.2 0 1 1 1.697 1.697l-2.758 3.152 2.758 3.15a1.2 1.2 0 0 1 0 1.698z"/>
					</svg>
				</button>
			</div>
		</div>
	</div>`

	return template.HTML(html)
}

// SuccessHTML generates HTML for success messages
func (h *ErrorHelpers) SuccessHTML(message string) template.HTML {
	if message == "" {
		return ""
	}

	html := `<div class="fixed top-4 right-4 z-50 bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded shadow-lg">
		<div class="flex">
			<div class="py-1">
				<svg class="fill-current h-6 w-6 text-green-500 mr-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20">
					<path d="M2.93 17.07A10 10 0 1 1 17.07 2.93 10 10 0 0 1 2.93 17.07zm12.73-1.41A8 8 0 1 0 4.34 4.34a8 8 0 0 0 11.32 11.32zM6.7 9.29L9 11.6l4.3-4.3 1.4 1.42L9 14.4l-3.7-3.7 1.4-1.41z"/>
				</svg>
			</div>
			<div>
				<p class="font-bold">Success</p>
				<p class="text-sm">` + template.HTMLEscapeString(message) + `</p>
			</div>
			<div class="ml-auto">
				<button type="button" class="text-green-700 hover:text-green-900" onclick="this.parentElement.parentElement.parentElement.remove()">
					<svg class="fill-current h-6 w-6" role="button" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20">
						<title>Close</title>
						<path d="M14.348 14.849a1.2 1.2 0 0 1-1.697 0L10 11.819l-2.651 3.029a1.2 1.2 0 1 1-1.697-1.697l2.758-3.15-2.759-3.152a1.2 1.2 0 1 1 1.697-1.697L10 8.183l2.651-3.031a1.2 1.2 0 1 1 1.697 1.697l-2.758 3.152 2.758 3.15a1.2 1.2 0 0 1 0 1.698z"/>
					</svg>
				</button>
			</div>
		</div>
	</div>`

	return template.HTML(html)
}