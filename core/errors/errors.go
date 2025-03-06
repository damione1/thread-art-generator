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
func UnauthenticatedError(err error) error {
	return status.Errorf(codes.Unauthenticated, "unauthorized: %s", err)
}

// RolePermissionError creates a gRPC PermissionDenied error
func RolePermissionError(err error) error {
	return status.Errorf(codes.PermissionDenied, "role permission error: %s", err)
}

// InternalError creates a gRPC Internal error
func InternalError(err error) error {
	return status.Errorf(codes.Internal, "internal error: %s", err)
}

// NotFoundError creates a gRPC NotFound error
func NotFoundError(err error) error {
	return status.Errorf(codes.NotFound, "not found: %s", err)
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
