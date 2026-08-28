package pbErrors

import (
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

// Error message constants
const (
	// Validation error prefix
	ErrValidationPrefix = "failed to validate request"

	// Common validation errors
	ErrEmailAlreadyExists        = "email already exists"
	ErrInvalidResourceName       = "invalid resource name"
	ErrUserNotFound              = "user not found"
	ErrIncorrectCredentials      = "incorrect email or password"
	ErrUserNotActive             = "user is not active"
	ErrTooManyValidationRequests = "too many validation requests"
	ErrSessionNotFound           = "session not found"
	ErrSessionBlocked            = "session is blocked"
)

// FieldViolation creates a field violation for gRPC error details
func FieldViolation(field string, err error) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       field,
		Description: err.Error(),
	}
}

func withBadRequest(code connect.Code, message string, violations []*errdetails.BadRequest_FieldViolation) error {
	cerr := connect.NewError(code, errors.New(message))
	if len(violations) == 0 {
		return cerr
	}
	detail, err := connect.NewErrorDetail(&errdetails.BadRequest{FieldViolations: violations})
	if err != nil {
		return cerr
	}
	cerr.AddDetail(detail)
	return cerr
}

// InvalidArgumentError creates an InvalidArgument error with field violations
func InvalidArgumentError(violations []*errdetails.BadRequest_FieldViolation) error {
	return withBadRequest(connect.CodeInvalidArgument, "invalid parameters", violations)
}

// UnauthenticatedError creates an Unauthenticated error
func UnauthenticatedError(message string) error {
	return connect.NewError(connect.CodeUnauthenticated, errors.New(message))
}

// PermissionDeniedError creates a PermissionDenied error
func PermissionDeniedError(message string) error {
	return connect.NewError(connect.CodePermissionDenied, errors.New(message))
}

// InternalError creates an Internal error
func InternalError(message string, err error) error {
	cerr := connect.NewError(connect.CodeInternal, errors.New(message))
	if err == nil {
		return cerr
	}
	detail, dErr := connect.NewErrorDetail(&errdetails.ErrorInfo{
		Reason: "INTERNAL_ERROR",
		Metadata: map[string]string{
			"error": err.Error(),
		},
	})
	if dErr != nil {
		return cerr
	}
	cerr.AddDetail(detail)
	return cerr
}

// NotFoundError creates a NotFound error
func NotFoundError(message string) error {
	return connect.NewError(connect.CodeNotFound, errors.New(message))
}

// AlreadyExistsError creates an AlreadyExists error
func AlreadyExistsError(message string, field string) error {
	if field == "" {
		return connect.NewError(connect.CodeAlreadyExists, errors.New(message))
	}
	return withBadRequest(connect.CodeAlreadyExists, message, []*errdetails.BadRequest_FieldViolation{
		FieldViolation(field, errors.New(message)),
	})
}

// FailedPreconditionError creates a FailedPrecondition error
func FailedPreconditionError(message string) error {
	return connect.NewError(connect.CodeFailedPrecondition, errors.New(message))
}

// FormatValidationError formats a validation error with the standard prefix
func FormatValidationError(err error) error {
	return fmt.Errorf("%s: %w", ErrValidationPrefix, err)
}

// NewValidationError creates a new validation error with the standard prefix
func NewValidationError(message string) error {
	return fmt.Errorf("%s: %s", ErrValidationPrefix, message)
}

// NewFieldValidationError creates a new field validation error with the standard format
func NewFieldValidationError(field, message string) error {
	return fmt.Errorf("%s: (%s: %s)", ErrValidationPrefix, field, message)
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string) error {
	return NotFoundError(message)
}

// NewInternalError creates a new internal error
func NewInternalError(message string, err error) error {
	return InternalError(message, err)
}

// NewUnauthenticatedError creates a new unauthenticated error
func NewUnauthenticatedError(message string) error {
	return UnauthenticatedError(message)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(fmt.Sprint(err), ErrValidationPrefix)
}

func hasCode(err error, code connect.Code) bool {
	return err != nil && connect.CodeOf(err) == code
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	return hasCode(err, connect.CodeNotFound)
}

// IsUnauthenticatedError checks if an error is an unauthenticated error
func IsUnauthenticatedError(err error) bool {
	return hasCode(err, connect.CodeUnauthenticated)
}

// IsPermissionDeniedError checks if an error is a permission denied error
func IsPermissionDeniedError(err error) bool {
	return hasCode(err, connect.CodePermissionDenied)
}

// IsInternalError checks if an error is an internal error
func IsInternalError(err error) bool {
	return hasCode(err, connect.CodeInternal)
}

// IsInvalidArgumentError checks if an error is an invalid argument error
func IsInvalidArgumentError(err error) bool {
	return hasCode(err, connect.CodeInvalidArgument)
}

// ExtractFieldViolations extracts field violations from an error
func ExtractFieldViolations(err error) []*errdetails.BadRequest_FieldViolation {
	if err == nil {
		return nil
	}
	var ce *connect.Error
	if !errors.As(err, &ce) {
		return nil
	}
	for _, d := range ce.Details() {
		msg, valErr := d.Value()
		if valErr != nil {
			continue
		}
		if badRequest, ok := msg.(*errdetails.BadRequest); ok {
			return badRequest.GetFieldViolations()
		}
	}
	return nil
}

// HasFieldViolation checks if an error has a field violation for a specific field
func HasFieldViolation(err error, field string) bool {
	violations := ExtractFieldViolations(err)
	for _, v := range violations {
		if v.GetField() == field {
			return true
		}
	}
	return false
}

// GetFieldViolationMessage gets the message for a specific field violation
func GetFieldViolationMessage(err error, field string) string {
	violations := ExtractFieldViolations(err)
	for _, v := range violations {
		if v.GetField() == field {
			return v.GetDescription()
		}
	}
	return ""
}

// ConvertProtoValidateError converts a protovalidate error to an InvalidArgumentError
func ConvertProtoValidateError(err error) error {
	if err == nil {
		return nil
	}

	validationErr, ok := err.(*protovalidate.ValidationError)
	if !ok {
		return InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			FieldViolation("", errors.New("validation failed")),
		})
	}

	fieldViolations := make([]*errdetails.BadRequest_FieldViolation, 0, len(validationErr.Violations))
	for _, violation := range validationErr.Violations {
		fieldPath := extractFieldName(violation)
		fieldViolations = append(fieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       fieldPath,
			Description: violation.Proto.GetMessage(),
		})
	}

	return InvalidArgumentError(fieldViolations)
}

func extractFieldName(violation *protovalidate.Violation) string {
	if violation == nil || violation.FieldDescriptor == nil {
		return ""
	}

	fieldName := string(violation.FieldDescriptor.Name())
	parts := strings.Split(fieldName, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(string(parts[i][0])) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}
