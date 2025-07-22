package pbErrors

import (
	"fmt"

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
)

// ErrorParser provides utilities for parsing and converting different error types
type ErrorParser struct{}

// NewErrorParser creates a new error parser instance
func NewErrorParser() *ErrorParser {
	return &ErrorParser{}
}

// ParseProtoValidationError converts protovalidate errors to StandardError
func (p *ErrorParser) ParseProtoValidationError(err error) *StandardError {
	if err == nil {
		return nil
	}

	validationErr, ok := err.(*protovalidate.ValidationError)
	if !ok {
		return NewGlobalError(ErrorTypeValidation, "validation failed")
	}

	builder := NewValidationErrorBuilder("validation failed")

	for _, violation := range validationErr.Violations {
		fieldPath := p.extractFieldName(violation)
		message := violation.Proto.GetMessage()

		// If no specific field, treat as global validation error
		if fieldPath == "" {
			return NewGlobalError(ErrorTypeValidation, message)
		}

		builder.AddField(fieldPath, message)
	}

	return builder.Build()
}

// ParseConnectError converts Connect-RPC errors to StandardError
func (p *ErrorParser) ParseConnectError(err error) *StandardError {
	if err == nil {
		return nil
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		return NewGlobalError(ErrorTypeInternal, err.Error())
	}

	// Map Connect error codes to our error types
	var errorType ErrorType
	switch connectErr.Code() {
	case connect.CodeInvalidArgument:
		errorType = ErrorTypeValidation
	case connect.CodeNotFound:
		errorType = ErrorTypeNotFound
	case connect.CodePermissionDenied:
		errorType = ErrorTypeForbidden
	case connect.CodeAlreadyExists:
		errorType = ErrorTypeConflict
	case connect.CodeUnauthenticated:
		errorType = ErrorTypeUnauthorized
	default:
		errorType = ErrorTypeInternal
	}

	builder := NewValidationErrorBuilder(connectErr.Message())

	// For now, we'll use a simple approach since Connect error details parsing is complex
	// This can be enhanced later to parse specific error details

	standardErr := builder.Build()
	standardErr.Type = errorType

	// If no field errors were found, make it a global error
	if len(standardErr.Fields) == 0 {
		standardErr.GlobalError = connectErr.Message()
	}

	return standardErr
}

// extractFieldName extracts the field name from a violation in a format suitable for frontend
func (p *ErrorParser) extractFieldName(violation *protovalidate.Violation) string {
	if violation == nil || violation.FieldDescriptor == nil {
		return ""
	}

	// Get the field name from the descriptor
	fieldName := string(violation.FieldDescriptor.Name())

	// Convert from snake_case to the format expected by your frontend
	// You can customize this logic based on your form field naming conventions
	return p.normalizeFieldName(fieldName)
}

// normalizeFieldName converts proto field names to frontend field names
// Customize this based on your frontend field naming conventions
func (p *ErrorParser) normalizeFieldName(fieldName string) string {
	// For now, keep the original field name
	// You can implement camelCase conversion or other transformations here
	return fieldName
}

// ParseBusinessLogicError creates a business logic error
func (p *ErrorParser) ParseBusinessLogicError(errorType ErrorType, message string, field string) *StandardError {
	if field != "" {
		// Field-specific business logic error
		builder := NewValidationErrorBuilder(message)
		builder.AddField(field, message)
		return builder.Build()
	}

	// Global business logic error
	return NewGlobalError(errorType, message)
}

// FormErrorResponse represents the structure for frontend error handling
type FormErrorResponse struct {
	Success     bool                `json:"success"`
	Message     string              `json:"message,omitempty"`
	FieldErrors map[string][]string `json:"field_errors,omitempty"`
	GlobalError string              `json:"global_error,omitempty"`
	ErrorType   string              `json:"error_type,omitempty"`
}

// ToFormErrorResponse converts StandardError to a form-friendly response
func (p *ErrorParser) ToFormErrorResponse(err *StandardError) *FormErrorResponse {
	if err == nil {
		return &FormErrorResponse{Success: true}
	}

	response := &FormErrorResponse{
		Success:   false,
		Message:   err.Message,
		ErrorType: string(err.Type),
	}

	if err.HasFieldErrors() {
		response.FieldErrors = err.Fields
	}

	if err.HasGlobalError() {
		response.GlobalError = err.GlobalError
	}

	return response
}

// FromFormData creates a validation error from form data
// This is useful when validating form submissions on the server side
func (p *ErrorParser) FromFormData(fieldErrors map[string][]string, globalMessage string) *StandardError {
	if len(fieldErrors) == 0 && globalMessage == "" {
		return nil
	}

	message := "validation failed"
	if globalMessage != "" {
		message = globalMessage
	}

	standardErr := NewStandardValidationError(message, fieldErrors)
	if globalMessage != "" {
		standardErr.GlobalError = globalMessage
	}

	return standardErr
}

// Common validation helpers that can be shared between client and server

// ValidateRequired checks if a field is required and not empty
func (p *ErrorParser) ValidateRequired(value string, fieldName string) []string {
	if value == "" {
		return []string{fieldName + " is required"}
	}
	return nil
}

// ValidateEmail checks if an email is valid
func (p *ErrorParser) ValidateEmail(email string, fieldName string) []string {
	if email == "" {
		return nil // Use ValidateRequired for required check
	}

	// Basic email validation - you can enhance this
	if len(email) < 3 || !contains(email, "@") || !contains(email, ".") {
		return []string{"Please enter a valid email address"}
	}
	return nil
}

// ValidateLength checks string length constraints
func (p *ErrorParser) ValidateLength(value string, fieldName string, min, max int) []string {
	var errors []string

	if min > 0 && len(value) < min {
		errors = append(errors, fmt.Sprintf("%s must be at least %d characters", fieldName, min))
	}

	if max > 0 && len(value) > max {
		errors = append(errors, fmt.Sprintf("%s must be no more than %d characters", fieldName, max))
	}

	return errors
}

// ValidateRange checks numeric range constraints
func (p *ErrorParser) ValidateRange(value int, fieldName string, min, max int) []string {
	var errors []string

	if value < min {
		errors = append(errors, fmt.Sprintf("%s must be at least %d", fieldName, min))
	}

	if value > max {
		errors = append(errors, fmt.Sprintf("%s must be no more than %d", fieldName, max))
	}

	return errors
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BuildValidationError is a convenience function for building validation errors
func BuildValidationError() *ValidationErrorBuilder {
	return NewValidationErrorBuilder("validation failed")
}
