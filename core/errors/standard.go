package pbErrors

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

// ErrorType is the BFF/UI category. Wire codes stay connect.Code.
type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "VALIDATION_ERROR"
	ErrorTypeNotFound     ErrorType = "NOT_FOUND"
	ErrorTypeUnauthorized ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden    ErrorType = "FORBIDDEN"
	ErrorTypeConflict     ErrorType = "CONFLICT"
	ErrorTypePrecondition ErrorType = "FAILED_PRECONDITION"
	ErrorTypeInternal     ErrorType = "INTERNAL_ERROR"
	ErrorTypeRateLimit    ErrorType = "RATE_LIMIT"
	ErrorTypeUnavailable  ErrorType = "UNAVAILABLE"
)

// StandardError is the decoded form of a Connect error for the BFF and templates.
// Services emit via InvalidArgumentError / NotFoundError / InternalError — not this type.
type StandardError struct {
	Type        ErrorType
	Message     string
	Fields      map[string][]string
	GlobalError string
}

func NewStandardValidationError(message string, fieldErrors map[string][]string) *StandardError {
	if fieldErrors == nil {
		fieldErrors = make(map[string][]string)
	}
	return &StandardError{
		Type:    ErrorTypeValidation,
		Message: message,
		Fields:  fieldErrors,
	}
}

func NewGlobalError(errorType ErrorType, message string) *StandardError {
	return &StandardError{
		Type:        errorType,
		Message:     message,
		GlobalError: message,
		Fields:      make(map[string][]string),
	}
}

func (e *StandardError) AddFieldError(field string, message string) {
	if e.Fields == nil {
		e.Fields = make(map[string][]string)
	}
	e.Fields[field] = append(e.Fields[field], message)
}

func (e *StandardError) HasFieldErrors() bool {
	return len(e.Fields) > 0
}

func (e *StandardError) HasGlobalError() bool {
	return e.GlobalError != ""
}

// ToConnectError is the inverse of FromConnectError (tests / ValidationErrorBuilder).
func (e *StandardError) ToConnectError() error {
	if e == nil {
		return nil
	}

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
	case ErrorTypePrecondition:
		code = connect.CodeFailedPrecondition
	case ErrorTypeRateLimit:
		code = connect.CodeResourceExhausted
	case ErrorTypeUnavailable:
		code = connect.CodeUnavailable
	default:
		code = connect.CodeInternal
	}

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
		return withBadRequest(code, e.Message, violations)
	}

	return connect.NewError(code, errors.New(e.Message))
}

// FromConnectError unwraps a Connect error (including fmt.Errorf %w wraps).
func FromConnectError(err error) *StandardError {
	if err == nil {
		return nil
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		return NewGlobalError(ErrorTypeInternal, UserFacingInternalMessage)
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
	case connect.CodeFailedPrecondition:
		errorType = ErrorTypePrecondition
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
	}

	for _, detail := range connectErr.Details() {
		value, valErr := detail.Value()
		if valErr != nil {
			continue
		}
		// BadRequest only. ErrorInfo used to carry SQL; never copy metadata to the UI.
		if badRequest, ok := value.(*errdetails.BadRequest); ok {
			for _, violation := range badRequest.FieldViolations {
				standardErr.AddFieldError(violation.Field, violation.Description)
			}
		}
	}

	if errorType == ErrorTypeValidation && !standardErr.HasFieldErrors() {
		standardErr.GlobalError = standardErr.Message
	} else if errorType == ErrorTypeInternal {
		standardErr.GlobalError = UserFacingInternalMessage
	} else if errorType != ErrorTypeValidation {
		standardErr.GlobalError = standardErr.Message
	}

	return standardErr
}

type ValidationErrorBuilder struct {
	message string
	fields  map[string][]string
}

func NewValidationErrorBuilder(message string) *ValidationErrorBuilder {
	return &ValidationErrorBuilder{
		message: message,
		fields:  make(map[string][]string),
	}
}

func (b *ValidationErrorBuilder) AddField(field, message string) *ValidationErrorBuilder {
	b.fields[field] = append(b.fields[field], message)
	return b
}

func (b *ValidationErrorBuilder) Build() *StandardError {
	return NewStandardValidationError(b.message, b.fields)
}

func (b *ValidationErrorBuilder) BuildConnectError() error {
	return b.Build().ToConnectError()
}

func BuildValidationError() *ValidationErrorBuilder {
	return NewValidationErrorBuilder("validation failed")
}

func (e *StandardError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}
