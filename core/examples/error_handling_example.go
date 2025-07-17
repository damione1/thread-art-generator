package examples

import (
	"context"
	"fmt"

	"github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/bufbuild/protovalidate-go"
)

// ExampleValidationWithStandardErrors demonstrates how to use the new standardized error handling
func ExampleValidationWithStandardErrors(ctx context.Context, req *pb.CreateArtRequest) (*pb.Art, error) {
	// Create error parser instance
	parser := pbErrors.NewErrorParser()

	// 1. Proto validation using standardized error handling
	if err := protovalidate.Validate(req); err != nil {
		standardErr := parser.ParseProtoValidationError(err)
		return nil, standardErr.ToConnectError()
	}

	// 2. Business logic validation with field-specific errors
	validationBuilder := pbErrors.NewValidationErrorBuilder("validation failed")
	
	// Check title length
	if req.Art.Title == "" {
		validationBuilder.AddField("art.title", "Title is required")
	} else if len(req.Art.Title) < 3 {
		validationBuilder.AddField("art.title", "Title must be at least 3 characters")
	} else if len(req.Art.Title) > 100 {
		validationBuilder.AddField("art.title", "Title must be no more than 100 characters")
	}

	// If we have validation errors, return them
	if len(validationBuilder.Build().Fields) > 0 {
		return nil, validationBuilder.BuildConnectError()
	}

	// 3. Business logic errors (non-field specific)
	// For example, checking if user has permission
	userID := "user-123" // This would come from context
	if userID == "" {
		return nil, pbErrors.StandardUnauthorizedError("user not authenticated")
	}

	// 4. Database/resource errors
	// For example, checking if art already exists
	existingArt := false // This would be a real database check
	if existingArt {
		return nil, pbErrors.StandardConflictError("art", "an art with this title already exists")
	}

	// Success case
	art := &pb.Art{
		Name:  fmt.Sprintf("users/%s/arts/new-art-id", userID),
		Title: req.Art.Title,
	}

	return art, nil
}

// ExampleClientErrorHandling demonstrates how to handle errors on the client side
func ExampleClientErrorHandling(err error) {
	if err == nil {
		fmt.Println("Success!")
		return
	}

	// Parse the error using standardized error handling
	standardErr := pbErrors.FromConnectError(err)
	
	if standardErr.HasFieldErrors() {
		fmt.Println("Field validation errors:")
		for field, messages := range standardErr.Fields {
			for _, message := range messages {
				fmt.Printf("  %s: %s\n", field, message)
			}
		}
	}

	if standardErr.HasGlobalError() {
		fmt.Printf("Global error: %s\n", standardErr.GlobalError)
	}

	// Or convert to form error response for frontend
	parser := pbErrors.NewErrorParser()
	formResponse := parser.ToFormErrorResponse(standardErr)
	
	if !formResponse.Success {
		fmt.Printf("Form error response: %+v\n", formResponse)
	}
}

// ExampleFormValidation demonstrates form validation helpers
func ExampleFormValidation(email, title string) *pbErrors.StandardError {
	parser := pbErrors.NewErrorParser()
	builder := pbErrors.NewValidationErrorBuilder("validation failed")

	// Validate email
	if emailErrors := parser.ValidateEmail(email, "Email"); len(emailErrors) > 0 {
		for _, err := range emailErrors {
			builder.AddField("email", err)
		}
	}

	// Validate title length
	if titleErrors := parser.ValidateLength(title, "Title", 3, 100); len(titleErrors) > 0 {
		for _, err := range titleErrors {
			builder.AddField("title", err)
		}
	}

	// Return validation errors if any
	validationErr := builder.Build()
	if len(validationErr.Fields) > 0 {
		return validationErr
	}

	return nil // No errors
}

// ExampleSharedValidation demonstrates how validation logic can be shared
func ExampleSharedValidation() {
	parser := pbErrors.NewErrorParser()

	// These validation functions can be used on both client and server
	titleErrors := parser.ValidateLength("Hi", "Title", 3, 100)
	emailErrors := parser.ValidateEmail("invalid-email", "Email")
	
	if len(titleErrors) > 0 || len(emailErrors) > 0 {
		builder := pbErrors.NewValidationErrorBuilder("validation failed")
		
		for _, err := range titleErrors {
			builder.AddField("title", err)
		}
		
		for _, err := range emailErrors {
			builder.AddField("email", err)
		}
		
		validationErr := builder.Build()
		fmt.Printf("Validation errors: %+v\n", validationErr.Fields)
	}
}