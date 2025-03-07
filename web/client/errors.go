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

	// Convert gRPC error to status
	st, ok := status.FromError(err)
	if !ok {
		result.GeneralError = err.Error()
		return result
	}

	// Extract field violations from the error details
	hasFieldViolations := false
	for _, detail := range st.Details() {
		switch d := detail.(type) {
		case *errdetails.BadRequest:
			// Process field violations
			if len(d.GetFieldViolations()) > 0 {
				hasFieldViolations = true
				for _, violation := range d.GetFieldViolations() {
					result.FieldErrors[violation.GetField()] = violation.GetDescription()
				}
			}
		}
	}

	// Only set the general error message if there are no field violations
	// This prevents "invalid parameters" from showing up when we have specific field errors
	if !hasFieldViolations {
		result.GeneralError = st.Message()
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
		// Parse the JSON into a status proto
		statusProto := status.New(0, "").Proto()
		err := protojson.Unmarshal([]byte(errorMessage), statusProto)

		if err == nil {
			hasFieldViolations := false

			// Extract field violations from details
			for _, detail := range statusProto.GetDetails() {
				// Parse the Any message
				var badRequest errdetails.BadRequest
				if detail.MessageIs(&badRequest) {
					if err := proto.Unmarshal(detail.Value, &badRequest); err == nil {
						if len(badRequest.GetFieldViolations()) > 0 {
							hasFieldViolations = true
							for _, violation := range badRequest.GetFieldViolations() {
								result.FieldErrors[violation.GetField()] = violation.GetDescription()
							}
						}
					}
				}
			}

			// Only set general error if no field violations
			if !hasFieldViolations {
				result.GeneralError = statusProto.GetMessage()
			}

			// If we have any field errors, return them
			if len(result.FieldErrors) > 0 {
				return result
			}
		} else {
			// Fallback to manual JSON parsing if protojson fails
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(errorMessage), &data); err == nil {
				hasFieldViolations := false

				// Try to extract details
				if details, ok := data["details"].([]interface{}); ok {
					for _, detail := range details {
						detailMap, ok := detail.(map[string]interface{})
						if !ok {
							continue
						}

						// Look for field violations
						violations, ok := detailMap["fieldViolations"].([]interface{})
						if !ok {
							continue
						}

						if len(violations) > 0 {
							hasFieldViolations = true
							for _, v := range violations {
								violation, ok := v.(map[string]interface{})
								if !ok {
									continue
								}

								field, _ := violation["field"].(string)
								description, _ := violation["description"].(string)

								if field != "" && description != "" {
									result.FieldErrors[field] = description
								}
							}
						}
					}
				}

				// Extract the message only if no field violations
				if !hasFieldViolations {
					if msg, ok := data["message"].(string); ok {
						result.GeneralError = msg
					}
				}
			}
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
