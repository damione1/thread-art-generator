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

The Go frontend parses these error messages using the `ParseValidationErrors` function in `pkg/errors/validation.go`:

```go
// FieldErrors maps field names to error messages
type FieldErrors map[string]string

// ParseValidationErrors parses a gRPC error message into field errors
func ParseValidationErrors(err error) FieldErrors {
	fieldErrors := make(FieldErrors)

	// Handle Connect-RPC errors
	if connectErr, ok := err.(*connect.Error); ok {
		// Check if it's a validation error (InvalidArgument code)
		if connectErr.Code() == connect.CodeInvalidArgument {
			// Extract details from error
			for _, detail := range connectErr.Details() {
				// Parse bad request details
				if br, ok := detail.Value.(*errdetails.BadRequest); ok {
					for _, violation := range br.GetFieldViolations() {
						fieldName := mapFieldName(violation.GetField())
						fieldErrors[fieldName] = violation.GetDescription()
					}
				}
			}

			// If no structured details, parse the error message
			if len(fieldErrors) == 0 {
				errMsg := connectErr.Message()
				parseErrorMessage(errMsg, fieldErrors)
			}
		}
	} else if err != nil {
		// Handle plain errors by parsing the message
		parseErrorMessage(err.Error(), fieldErrors)
	}

	return fieldErrors
}

// parseErrorMessage parses error messages in our standard format
func parseErrorMessage(errorMessage string, fieldErrors FieldErrors) {
	// Check if it's a validation error
	if !strings.Contains(errorMessage, "failed to validate request") {
		// Not a validation error, add a general error
		fieldErrors["general"] = errorMessage
		return
	}

	// Parse field errors in the format "(field: message; field2: message2)"
	fieldErrorsRegex := regexp.MustCompile(`\(([^)]+)\)`)
	matches := fieldErrorsRegex.FindStringSubmatch(errorMessage)

	if len(matches) > 1 {
		// Process field errors
		fields := strings.Split(matches[1], ";")
		for _, field := range fields {
			parts := strings.SplitN(strings.TrimSpace(field), ":", 2)
			if len(parts) == 2 {
				fieldName := mapFieldName(strings.TrimSpace(parts[0]))
				fieldErrors[fieldName] = strings.TrimSpace(parts[1])
			}
		}
	} else {
		// Handle simple validation errors without field information
		simpleMsgRegex := regexp.MustCompile(`failed to validate request: (.+)`)
		simpleMatches := simpleMsgRegex.FindStringSubmatch(errorMessage)
		if len(simpleMatches) > 1 {
			fieldErrors["general"] = simpleMatches[1]
		} else {
			fieldErrors["general"] = errorMessage
		}
	}
}

// mapFieldName maps backend field names to frontend field names
func mapFieldName(fieldName string) string {
	fieldMapping := map[string]string{
		"first_name": "firstName",
		"last_name":  "lastName",
		// Add more mappings as needed
	}

	if mapped, ok := fieldMapping[fieldName]; ok {
		return mapped
	}
	return fieldName
}
```

### Using Error Handling in Templates

In your HTML templates, use the parsed errors to show field-level validation messages:

```html
{{ define "form-field" }}
<div class="form-group {{ if index .Errors .Name }}has-error{{ end }}">
  <label for="{{ .Name }}">{{ .Label }}</label>
  <input
    type="{{ .Type }}"
    id="{{ .Name }}"
    name="{{ .Name }}"
    value="{{ .Value }}"
    class="form-control"
  />
  {{ if index .Errors .Name }}
  <div class="error-message">{{ index .Errors .Name }}</div>
  {{ end }}
</div>
{{ end }}
```

And for general errors:

```html
{{ if index .Errors "general" }}
<div class="alert alert-danger">{{ index .Errors "general" }}</div>
{{ end }}
```

## Best Practices

1. **Always use the helper functions** for creating errors in the backend
2. **Use the error constants** instead of string literals
3. **Follow the standard error formats** for consistency
4. **Add new error constants** to `errors.go` when needed
5. **Update the frontend error parser** when adding new error formats
6. **Use error type checking** functions to handle specific error types

## Testing

Ensure all error formats are covered in the tests for `ParseValidationErrors` in `pkg/errors/validation_test.go`.
