import { ConnectError, Code } from "@connectrpc/connect";
import { BadRequest } from "./pb/google/rpc/error_details_pb";

// Helper type for field errors
export interface FieldErrors {
    [fieldName: string]: string;
}

/**
 * Extract field violations from a gRPC error
 * @param error The gRPC error
 * @returns An object with field names as keys and error messages as values
 */
export function extractFieldErrors(error: unknown): FieldErrors {
    if (!(error instanceof ConnectError)) {
        return {};
    }

    // Only process InvalidArgument errors
    if (error.code !== Code.InvalidArgument) {
        return {};
    }

    const fieldErrors: FieldErrors = {};

    try {
        // Get the first BadRequest detail from the error
        const badRequestDetails = error.findDetails(BadRequest);
        if (badRequestDetails.length > 0) {
            const badRequest = badRequestDetails[0];

            // Process each field violation
            for (const violation of badRequest.fieldViolations) {
                // Map the field path to the corresponding form field
                const fieldName = mapFieldPathToFormField(violation.field);
                fieldErrors[fieldName] = violation.description;
            }
        }
    } catch (err) {
        console.info("Error parsing field violations:", err);
    }

    return fieldErrors;
}

/**
 * Map a protobuf field path to a form field name
 * This may need customization based on your specific field naming conventions
 */
function mapFieldPathToFormField(fieldPath: string): string {
    // Handle nested fields (e.g., "user.email" -> "email")
    const parts = fieldPath.split('.');
    if (parts.length > 1) {
        return parts[parts.length - 1];
    }

    // Handle array indices (e.g., "emails[0]" -> "emails")
    const match = fieldPath.match(/^([^\[]+)(\[\d+\])?(.*)$/);
    if (match) {
        return match[1] + (match[3] || '');
    }

    return fieldPath;
}

/**
 * Centralized error handler for form submissions
 * @param error The error from the gRPC call
 * @param setFieldError Function to set a field-specific error
 * @param setGeneralError Function to set a general form error
 * @returns True if the error was handled, false otherwise
 */
export function handleFormError(
    error: unknown,
    setFieldError: (field: string, error: string) => void,
    setGeneralError: (error: string | null) => void
): boolean {
    if (!(error instanceof ConnectError)) {
        setGeneralError(error instanceof Error ? error.message : "An unknown error occurred");
        return true;
    }

    // Handle different gRPC error codes
    switch (error.code) {
        case Code.InvalidArgument:
            // Extract and set field errors
            const fieldErrors = extractFieldErrors(error);
            if (Object.keys(fieldErrors).length > 0) {
                // Set individual field errors
                Object.entries(fieldErrors).forEach(([field, message]) => {
                    setFieldError(field, message);
                });
                setGeneralError("Please correct the errors in the form.");
                return true;
            }
            setGeneralError(error.message);
            return true;

        case Code.Unauthenticated:
            setGeneralError("Authentication required. Please log in again.");
            return true;

        case Code.PermissionDenied:
            setGeneralError("You don't have permission to perform this action.");
            return true;

        case Code.AlreadyExists:
            // Try to extract field from the error if available
            const fieldErrors2 = extractFieldErrors(error);
            if (Object.keys(fieldErrors2).length > 0) {
                Object.entries(fieldErrors2).forEach(([field, message]) => {
                    setFieldError(field, message);
                });
                return true;
            }
            setGeneralError(error.message || "This resource already exists.");
            return true;

        case Code.Internal:
            setGeneralError("An internal server error occurred. Please try again later.");
            return true;

        default:
            setGeneralError(error.message || "An error occurred.");
            return true;
    }
}
