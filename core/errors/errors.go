package pbErrors

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// InvalidArgumentError creates a gRPC InvalidArgument error with field violations
func InvalidArgumentError(violations []*errdetails.BadRequest_FieldViolation) error {
	badRequest := &errdetails.BadRequest{FieldViolations: violations}
	statusInvalid := status.New(codes.InvalidArgument, "invalid parameters")

	statusDetails, err := statusInvalid.WithDetails(badRequest)
	if err != nil {
		return statusInvalid.Err()
	}

	return statusDetails.Err()
}

// UnauthenticatedError creates a gRPC Unauthenticated error
func UnauthenticatedError(message string) error {
	st := status.New(codes.Unauthenticated, message)
	return st.Err()
}

// PermissionDeniedError creates a gRPC PermissionDenied error
func PermissionDeniedError(message string) error {
	st := status.New(codes.PermissionDenied, message)
	return st.Err()
}

// InternalError creates a gRPC Internal error
func InternalError(message string, err error) error {
	st := status.New(codes.Internal, message)

	if err != nil {
		// Add error details
		errorInfo := &errdetails.ErrorInfo{
			Reason: "INTERNAL_ERROR",
			Metadata: map[string]string{
				"error": err.Error(),
			},
		}

		statusWithDetails, detailErr := st.WithDetails(errorInfo)
		if detailErr != nil {
			return st.Err()
		}
		return statusWithDetails.Err()
	}

	return st.Err()
}

// NotFoundError creates a gRPC NotFound error
func NotFoundError(message string) error {
	st := status.New(codes.NotFound, message)
	return st.Err()
}

// AlreadyExistsError creates a gRPC AlreadyExists error
func AlreadyExistsError(message string, field string) error {
	st := status.New(codes.AlreadyExists, message)

	if field != "" {
		violation := FieldViolation(field, errors.New(message))
		badRequest := &errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{violation},
		}

		statusWithDetails, detailErr := st.WithDetails(badRequest)
		if detailErr != nil {
			return st.Err()
		}
		return statusWithDetails.Err()
	}

	return st.Err()
}

// FailedPreconditionError creates a gRPC FailedPrecondition error
func FailedPreconditionError(message string) error {
	st := status.New(codes.FailedPrecondition, message)
	return st.Err()
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

// NewNotFoundError creates a new not found error with the standard gRPC status
func NewNotFoundError(message string) error {
	return status.Errorf(codes.NotFound, message)
}

// NewInternalError creates a new internal error with the standard gRPC status
func NewInternalError(message string, err error) error {
	return status.Errorf(codes.Internal, "%s: %s", message, err)
}

// NewUnauthenticatedError creates a new unauthenticated error with the standard gRPC status
func NewUnauthenticatedError(message string) error {
	return status.Errorf(codes.Unauthenticated, message)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, errors.New(ErrValidationPrefix)) ||
		strings.Contains(fmt.Sprint(err), ErrValidationPrefix)
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	s, ok := status.FromError(err)
	return ok && s.Code() == codes.NotFound
}

// IsUnauthenticatedError checks if an error is an unauthenticated error
func IsUnauthenticatedError(err error) bool {
	if err == nil {
		return false
	}
	s, ok := status.FromError(err)
	return ok && s.Code() == codes.Unauthenticated
}

// IsPermissionDeniedError checks if an error is a permission denied error
func IsPermissionDeniedError(err error) bool {
	if err == nil {
		return false
	}
	s, ok := status.FromError(err)
	return ok && s.Code() == codes.PermissionDenied
}

// IsInternalError checks if an error is an internal error
func IsInternalError(err error) bool {
	if err == nil {
		return false
	}
	s, ok := status.FromError(err)
	return ok && s.Code() == codes.Internal
}

// IsInvalidArgumentError checks if an error is an invalid argument error
func IsInvalidArgumentError(err error) bool {
	if err == nil {
		return false
	}
	s, ok := status.FromError(err)
	return ok && s.Code() == codes.InvalidArgument
}

// ExtractFieldViolations extracts field violations from an error
func ExtractFieldViolations(err error) []*errdetails.BadRequest_FieldViolation {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return nil
	}

	for _, detail := range st.Details() {
		if badRequest, ok := detail.(*errdetails.BadRequest); ok {
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
