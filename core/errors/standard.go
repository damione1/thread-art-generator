package pbErrors

import (
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

// ErrorType represents the category of error for consistent handling
type ErrorType string

const (
	// Field-level errors that should be mapped to specific form fields
	ErrorTypeValidation ErrorType = "VALIDATION_ERROR"

	// Global errors that should be shown as toasts or general messages
	ErrorTypeNotFound     ErrorType = "NOT_FOUND"
	ErrorTypeUnauthorized ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden    ErrorType = "FORBIDDEN"
	ErrorTypeConflict     ErrorType = "CONFLICT"
	ErrorTypeInternal     ErrorType = "INTERNAL_ERROR"
	ErrorTypeRateLimit    ErrorType = "RATE_LIMIT"
	ErrorTypeUnavailable  ErrorType = "UNAVAILABLE"
)

// StandardError represents a standardized error structure
type StandardError struct {
	Type        ErrorType
	Message     string
	Fields      map[string][]string // Field-specific error messages
	GlobalError string              // Global error message for toasts
	Details     map[string]string   // Additional error details
}

// NewStandardValidationError creates a new validation error with field mappings
func NewStandardValidationError(message string, fieldErrors map[string][]string) *StandardError {
	return &StandardError{
		Type:    ErrorTypeValidation,
		Message: message,
		Fields:  fieldErrors,
		Details: make(map[string]string),
	}
}

// NewGlobalError creates a new global error (for toast notifications)
func NewGlobalError(errorType ErrorType, message string) *StandardError {
	return &StandardError{
		Type:        errorType,
		Message:     message,
		GlobalError: message,
		Fields:      make(map[string][]string),
		Details:     make(map[string]string),
	}
}

// AddFieldError adds a field-specific error
func (e *StandardError) AddFieldError(field string, message string) {
	if e.Fields == nil {
		e.Fields = make(map[string][]string)
	}
	e.Fields[field] = append(e.Fields[field], message)
}

// AddDetail adds additional error details
func (e *StandardError) AddDetail(key, value string) {
	if e.Details == nil {
		e.Details = make(map[string]string)
	}
	e.Details[key] = value
}

// HasFieldErrors returns true if there are field-specific errors
func (e *StandardError) HasFieldErrors() bool {
	return len(e.Fields) > 0
}

// HasGlobalError returns true if there's a global error message
func (e *StandardError) HasGlobalError() bool {
	return e.GlobalError != ""
}

// ToConnectError converts StandardError to a Connect-RPC error
func (e *StandardError) ToConnectError() error {
	var code connect.Code

	switch e.Type {
	case ErrorTypeValidation:
		code = connect.CodeInvalidArgument
	case ErrorTypeNotFound:
		code = connect.CodeNotFound
	case ErrorTypeUnauthorized:
		code = connect.CodeUnauthenticated
	case ErrorTypeForbidden:
		code = connect.CodePermissionDenied
	case ErrorTypeConflict:
		code = connect.CodeAlreadyExists
	case ErrorTypeRateLimit:
		code = connect.CodeResourceExhausted
	case ErrorTypeUnavailable:
		code = connect.CodeUnavailable
	default:
		code = connect.CodeInternal
	}

	connectErr := connect.NewError(code, fmt.Errorf("%s", e.Message))

	// Add field violations if present
	if e.HasFieldErrors() {
		var violations []*errdetails.BadRequest_FieldViolation
		for field, messages := range e.Fields {
			for _, msg := range messages {
				violations = append(violations, &errdetails.BadRequest_FieldViolation{
					Field:       field,
					Description: msg,
				})
			}
		}

		badRequest := &errdetails.BadRequest{FieldViolations: violations}
		if detail, err := connect.NewErrorDetail(badRequest); err == nil {
			connectErr.AddDetail(detail)
		}
	}

	// Add error info for additional details
	if len(e.Details) > 0 {
		errorInfo := &errdetails.ErrorInfo{
			Reason:   string(e.Type),
			Metadata: e.Details,
		}
		if detail, err := connect.NewErrorDetail(errorInfo); err == nil {
			connectErr.AddDetail(detail)
		}
	}

	return connectErr
}

// FromConnectError converts a Connect-RPC error to StandardError
func FromConnectError(err error) *StandardError {
	if err == nil {
		return nil
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		return NewGlobalError(ErrorTypeInternal, err.Error())
	}

	var errorType ErrorType
	switch connectErr.Code() {
	case connect.CodeInvalidArgument:
		errorType = ErrorTypeValidation
	case connect.CodeNotFound:
		errorType = ErrorTypeNotFound
	case connect.CodeUnauthenticated:
		errorType = ErrorTypeUnauthorized
	case connect.CodePermissionDenied:
		errorType = ErrorTypeForbidden
	case connect.CodeAlreadyExists:
		errorType = ErrorTypeConflict
	case connect.CodeResourceExhausted:
		errorType = ErrorTypeRateLimit
	case connect.CodeUnavailable:
		errorType = ErrorTypeUnavailable
	default:
		errorType = ErrorTypeInternal
	}

	standardErr := &StandardError{
		Type:    errorType,
		Message: connectErr.Message(),
		Fields:  make(map[string][]string),
		Details: make(map[string]string),
	}

	// Extract field violations and error info from details
	for _, detail := range connectErr.Details() {
		value, err := detail.Value()
		if err != nil {
			continue
		}
		switch d := value.(type) {
		case *errdetails.BadRequest:
			for _, violation := range d.FieldViolations {
				standardErr.AddFieldError(violation.Field, violation.Description)
			}
		case *errdetails.ErrorInfo:
			for key, value := range d.Metadata {
				standardErr.AddDetail(key, value)
			}
		}
	}

	// If it's a validation error but no field errors, treat as global
	if errorType == ErrorTypeValidation && !standardErr.HasFieldErrors() {
		standardErr.GlobalError = standardErr.Message
	} else if errorType != ErrorTypeValidation {
		// Non-validation errors are always global
		standardErr.GlobalError = standardErr.Message
	}

	return standardErr
}

// ValidationErrorBuilder helps build validation errors consistently
type ValidationErrorBuilder struct {
	message string
	fields  map[string][]string
}

// NewValidationErrorBuilder creates a new validation error builder
func NewValidationErrorBuilder(message string) *ValidationErrorBuilder {
	return &ValidationErrorBuilder{
		message: message,
		fields:  make(map[string][]string),
	}
}

// AddField adds a field error to the builder
func (b *ValidationErrorBuilder) AddField(field, message string) *ValidationErrorBuilder {
	b.fields[field] = append(b.fields[field], message)
	return b
}

// Build creates the final StandardError
func (b *ValidationErrorBuilder) Build() *StandardError {
	return NewStandardValidationError(b.message, b.fields)
}

// BuildConnectError creates a Connect-RPC error directly
func (b *ValidationErrorBuilder) BuildConnectError() error {
	return b.Build().ToConnectError()
}

// Common error builders for consistent error messages
func StandardNotFoundError(resource string) error {
	return NewGlobalError(ErrorTypeNotFound, fmt.Sprintf("%s not found", resource)).ToConnectError()
}

func StandardUnauthorizedError(message string) error {
	if message == "" {
		message = "Authentication required"
	}
	return NewGlobalError(ErrorTypeUnauthorized, message).ToConnectError()
}

func StandardForbiddenError(message string) error {
	if message == "" {
		message = "Access denied"
	}
	return NewGlobalError(ErrorTypeForbidden, message).ToConnectError()
}

func StandardConflictError(resource, reason string) error {
	message := fmt.Sprintf("%s already exists", resource)
	if reason != "" {
		message = fmt.Sprintf("%s: %s", message, reason)
	}
	return NewGlobalError(ErrorTypeConflict, message).ToConnectError()
}

func StandardInternalServerError(message string) error {
	if message == "" {
		message = "Internal server error"
	}
	return NewGlobalError(ErrorTypeInternal, message).ToConnectError()
}
