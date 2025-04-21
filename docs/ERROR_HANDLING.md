# Error Handling Guidelines

This document outlines the standardized approach to error handling in our application, ensuring consistency between backend and frontend.

## Backend Error Handling

### Error Constants

We use standardized error constants defined in `core/errors/errors.go` to ensure consistency:

```go
// Error message constants
const (
    // Validation error prefix
    ErrValidationPrefix = "failed to validate request"

    // Common validation errors
    ErrEmailAlreadyExists = "email already exists"
    ErrInvalidResourceName = "invalid resource name"
    ErrUserNotFound = "user not found"
    ErrIncorrectCredentials = "incorrect email or password"
    ErrUserNotActive = "user is not active"
    ErrTooManyValidationRequests = "too many validation requests"
    ErrSessionNotFound = "session not found"
    ErrSessionBlocked = "session is blocked"
)
```

### Helper Functions

We provide helper functions to create standardized errors:

```go
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
```

### gRPC Error Handling

For gRPC-specific errors, we provide additional helper functions:

```go
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

// NotFoundError creates a gRPC NotFound error
func NotFoundError(err error) error {
    return status.Errorf(codes.NotFound, "not found: %s", err)
}
```

### Error Type Checking

We also provide helper functions to check error types:

```go
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
```

### Error Formats

Our backend returns errors in several standardized formats:

1. Field validation errors:

   ```
   "failed to validate request: user: (email: cannot be blank; first_name: cannot be blank)"
   ```

2. Simple validation errors:

   ```
   "failed to validate request: email already exists"
   ```

3. Resource validation errors:

   ```
   "failed to validate request: invalid resource name: ..."
   ```

4. Authentication errors:
   ```
   "failed to validate request: (email: incorrect email or password; password: incorrect email or password)"
   ```

## Frontend Error Handling

The frontend parses these error messages using the `parseValidationErrors` function in `web/src/utils/errorUtils.ts`:

```typescript
export function parseValidationErrors(errorMessage: string): {
  [key: string]: string;
} {
  const errors: { [key: string]: string } = {};

  // Check if it's a validation error
  if (!errorMessage || !errorMessage.includes("failed to validate request")) {
    return errors;
  }

  // ... parsing logic ...

  return errors;
}
```

### Field Name Mapping

The frontend maps backend field names to frontend field names:

```typescript
const fieldMapping: { [key: string]: string } = {
  first_name: "firstName",
  last_name: "lastName",
  validation_number: "validationNumber",
  refresh_token: "refreshToken",
};
```

## Best Practices

1. **Always use the helper functions** for creating errors in the backend
2. **Use the error constants** instead of string literals
3. **Follow the standard error formats** for consistency
4. **Add new error constants** to `errors.go` when needed
5. **Update the frontend error parser** when adding new error formats
6. **Use error type checking** functions to handle specific error types

## Testing

Ensure all error formats are covered in the tests for `parseValidationErrors` in `web/src/utils/errorUtils.test.ts`.
