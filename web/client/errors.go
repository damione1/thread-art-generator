package client

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// FieldError represents an error for a specific field
type FieldError struct {
	Field   string
	Message string
}

// ValidationErrors holds all validation errors
type ValidationErrors struct {
	FieldErrors  map[string]string
	GeneralError string
}

// HasErrors returns true if there are any errors
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.FieldErrors) > 0 || ve.GeneralError != ""
}

// HasFieldError returns true if the specified field has an error
func (ve *ValidationErrors) HasFieldError(field string) bool {
	_, exists := ve.FieldErrors[field]
	return exists
}

// GetFieldError returns the error message for a field
func (ve *ValidationErrors) GetFieldError(field string) string {
	if message, exists := ve.FieldErrors[field]; exists {
		return message
	}
	return ""
}

// ExtractErrorDetails extracts detailed error information from a gRPC error
func ExtractErrorDetails(err error) *ValidationErrors {
	result := &ValidationErrors{
		FieldErrors: make(map[string]string),
	}

	if err == nil {
		return result
	}

	// Debug logging
	fmt.Printf("ExtractErrorDetails - Original error: %v\n", err)

	// Convert gRPC error to status
	st, ok := status.FromError(err)
	if !ok {
		fmt.Printf("ExtractErrorDetails - Not a gRPC status error\n")
		result.GeneralError = err.Error()
		return result
	}

	fmt.Printf("ExtractErrorDetails - gRPC status: code=%d, message=%s\n",
		st.Code(), st.Message())

	// Print all details for debugging
	fmt.Printf("ExtractErrorDetails - Status details count: %d\n", len(st.Details()))
	for i, detail := range st.Details() {
		fmt.Printf("ExtractErrorDetails - Detail #%d: %T, %+v\n", i, detail, detail)
	}

	// Extract field violations from the error details
	hasFieldViolations := false
	for _, detail := range st.Details() {
		switch d := detail.(type) {
		case *errdetails.BadRequest:
			fmt.Printf("ExtractErrorDetails - Found BadRequest detail with %d violations\n",
				len(d.GetFieldViolations()))

			// Process field violations
			if len(d.GetFieldViolations()) > 0 {
				hasFieldViolations = true
				for _, violation := range d.GetFieldViolations() {
					field := violation.GetField()
					description := violation.GetDescription()

					fmt.Printf("ExtractErrorDetails - Field violation: %s = %s\n", field, description)
					result.FieldErrors[field] = description
				}
			}
		default:
			fmt.Printf("ExtractErrorDetails - Unhandled detail type: %T\n", d)
		}
	}

	// Only set the general error message if there are no field violations
	// This prevents "invalid parameters" from showing up when we have specific field errors
	if !hasFieldViolations {
		result.GeneralError = st.Message()
		fmt.Printf("ExtractErrorDetails - Setting general error: %s\n", result.GeneralError)
	}

	return result
}

// ParseGRPCError converts a gRPC error to a JSON string for parsing by the frontend
func ParseGRPCError(err error) string {
	if err == nil {
		return ""
	}

	// Try to convert to gRPC status
	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC status error, return the error message
		return err.Error()
	}

	// Convert the proto to JSON using protojson
	jsonBytes, err := protojson.Marshal(st.Proto())
	if err != nil {
		return st.Message()
	}

	return string(jsonBytes)
}

// ParseValidationError parses a JSON error message from the API and returns structured validation errors
func ParseValidationError(errorMessage string) *ValidationErrors {
	result := &ValidationErrors{
		FieldErrors: make(map[string]string),
	}

	// If no error, return empty result
	if errorMessage == "" {
		return result
	}

	// Check if this is a gRPC error with details in JSON format
	if strings.Contains(errorMessage, "\"details\"") {
		fmt.Printf("ParseValidationError - Found details in JSON: %s\n", errorMessage)

		// Try direct JSON parsing first for best reliability
		var data map[string]interface{}
		err := json.Unmarshal([]byte(errorMessage), &data)
		if err == nil {
			fmt.Printf("ParseValidationError - Successfully parsed JSON\n")

			// Set general error message
			if msg, ok := data["message"].(string); ok {
				result.GeneralError = msg
				fmt.Printf("ParseValidationError - Found message: %s\n", msg)
			}

			// Look for field violations in details
			hasFieldViolations := false
			if details, ok := data["details"].([]interface{}); ok {
				fmt.Printf("ParseValidationError - Found %d details\n", len(details))

				for _, detail := range details {
					detailMap, ok := detail.(map[string]interface{})
					if !ok {
						continue
					}

					// Check for BadRequest type
					typeURL, _ := detailMap["@type"].(string)
					fmt.Printf("ParseValidationError - Detail type: %s\n", typeURL)

					if typeURL == "type.googleapis.com/google.rpc.BadRequest" {
						// Look for field violations
						violations, ok := detailMap["fieldViolations"].([]interface{})
						if !ok || len(violations) == 0 {
							fmt.Printf("ParseValidationError - No field violations found\n")
							continue
						}

						fmt.Printf("ParseValidationError - Found %d field violations\n", len(violations))
						hasFieldViolations = true

						for _, v := range violations {
							violation, ok := v.(map[string]interface{})
							if !ok {
								continue
							}

							field, _ := violation["field"].(string)
							description, _ := violation["description"].(string)

							fmt.Printf("ParseValidationError - Violation: %s = %s\n", field, description)

							if field != "" && description != "" {
								result.FieldErrors[field] = description
							}
						}
					}
				}
			}

			// Don't set general error if we have field violations
			if hasFieldViolations {
				result.GeneralError = ""
			}

			// If we have field errors, return the result
			if len(result.FieldErrors) > 0 {
				return result
			}
		}

		// If JSON parsing failed or didn't extract field violations, try with protojson
		fmt.Printf("ParseValidationError - Trying with protojson\n")
		statusProto := status.New(0, "").Proto()
		err = protojson.Unmarshal([]byte(errorMessage), statusProto)
		if err == nil {
			fmt.Printf("ParseValidationError - protojson parsing succeeded\n")

			hasFieldViolations := false
			// Extract field violations from details
			for _, detail := range statusProto.GetDetails() {
				fmt.Printf("ParseValidationError - Examining detail: %T\n", detail)
				// Parse the Any message
				var badRequest errdetails.BadRequest
				if detail.MessageIs(&badRequest) {
					fmt.Printf("ParseValidationError - Found BadRequest message\n")
					if err := proto.Unmarshal(detail.Value, &badRequest); err == nil {
						fmt.Printf("ParseValidationError - Unmarshaled BadRequest with %d violations\n",
							len(badRequest.GetFieldViolations()))
						if len(badRequest.GetFieldViolations()) > 0 {
							hasFieldViolations = true
							for _, violation := range badRequest.GetFieldViolations() {
								field := violation.GetField()
								description := violation.GetDescription()

								fmt.Printf("ParseValidationError - Field violation: %s = %s\n",
									field, description)

								result.FieldErrors[field] = description
							}
						}
					} else {
						fmt.Printf("ParseValidationError - Error unmarshaling BadRequest: %v\n", err)
					}
				}
			}

			// Only set general error if no field violations
			if !hasFieldViolations {
				result.GeneralError = statusProto.GetMessage()
				fmt.Printf("ParseValidationError - Setting general error: %s\n", result.GeneralError)
			} else {
				result.GeneralError = ""
			}

			// If we have any field errors, return them
			if len(result.FieldErrors) > 0 {
				return result
			}
		} else {
			fmt.Printf("ParseValidationError - protojson parsing failed: %v\n", err)
		}
	} else if strings.Contains(errorMessage, "failed to validate request") {
		// Legacy format handling - try to parse it
		parts := strings.SplitN(errorMessage, "failed to validate request: ", 2)
		if len(parts) >= 2 {
			errorBody := parts[1]

			// Extract content inside parentheses if present
			startIdx := strings.Index(errorBody, "(")
			endIdx := strings.LastIndex(errorBody, ")")

			if startIdx >= 0 && endIdx > startIdx {
				// We found content inside parentheses, parse it for field errors
				fieldViolations := errorBody[startIdx+1 : endIdx]

				// Split by semicolon to get individual field errors
				fieldErrors := strings.Split(fieldViolations, ";")
				hasFieldErrors := false

				for _, fieldError := range fieldErrors {
					fieldError = strings.TrimSpace(fieldError)
					if fieldError == "" {
						continue
					}

					// Split by colon to get field and message
					fieldParts := strings.SplitN(fieldError, ":", 2)
					if len(fieldParts) == 2 {
						hasFieldErrors = true
						field := strings.TrimSpace(fieldParts[0])
						message := strings.TrimSpace(fieldParts[1])
						result.FieldErrors[field] = message
					}
				}

				// Don't set general error if we have field errors
				if !hasFieldErrors {
					result.GeneralError = errorBody
				}
			} else {
				// No parentheses, use as general error
				result.GeneralError = errorBody
			}
		} else {
			// Couldn't split properly, use as general error
			result.GeneralError = errorMessage
		}
	} else {
		// Not a recognized format, use as general error
		result.GeneralError = errorMessage
	}

	return result
}

// Format creates a user-friendly error message for display
func (ve *ValidationErrors) Format() string {
	if ve.GeneralError != "" {
		return ve.GeneralError
	}

	if len(ve.FieldErrors) == 0 {
		return ""
	}

	// Format field errors as a list
	var messages []string
	for field, message := range ve.FieldErrors {
		messages = append(messages, fmt.Sprintf("%s: %s", field, message))
	}

	return strings.Join(messages, ", ")
}
