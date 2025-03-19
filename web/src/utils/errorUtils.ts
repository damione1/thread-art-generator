/**
 * Error utilities for handling backend validation errors
 */

type FieldErrors = {
    [key: string]: string;
};

/**
 * Field name mapping from backend format to frontend format
 */
const fieldMapping: { [key: string]: string } = {
    first_name: "firstName",
    last_name: "lastName",
    validation_number: "validationNumber",
    refresh_token: "refreshToken",
    // Add more mappings as needed
};

/**
 * Parses legacy validation errors following the format:
 * "failed to validate request: (field: error message)"
 */
export function parseValidationErrors(errorMessage: string): FieldErrors {
    const errors: FieldErrors = {};

    // Check if it's a validation error
    if (!errorMessage || !errorMessage.includes("failed to validate request")) {
        return errors;
    }

    // Extract field errors within parentheses
    const parenthesesMatch = errorMessage.match(/\(([^)]+)\)/);
    if (parenthesesMatch && parenthesesMatch[1]) {
        const fieldErrorPairs = parenthesesMatch[1].split(";");

        fieldErrorPairs.forEach((pair) => {
            const [field, message] = pair.split(":").map((s) => s.trim());
            if (field && message) {
                // Map field name if mapping exists
                const mappedField = fieldMapping[field] || field;
                errors[mappedField] = message;
            }
        });

        return errors;
    }

    // Simple validation error without field specification
    if (errorMessage.includes("failed to validate request: ")) {
        const simpleError = errorMessage.split("failed to validate request: ")[1].trim();
        // Use a special key for general errors
        errors["_general"] = simpleError;
    }

    return errors;
}

/**
 * Parses structured validation errors from protovalidate-go following the format:
 * "Error: [unknown] validation error: - user.last_name: value does not match regex pattern `^[a-zA-Z \-\']+$` [string.pattern]"
 */
export function parseProtoValidateErrors(errorMessage: string): FieldErrors {
    const errors: FieldErrors = {};

    if (!errorMessage || !errorMessage.includes("validation error:")) {
        return errors;
    }

    // Split the error message into lines to handle multiple validation errors
    const errorLines = errorMessage.split("- ").filter(line => line.trim() !== "");

    // Process each validation error line
    errorLines.forEach(line => {
        // Skip the prefix for the first line
        if (line.includes("Error: [unknown] validation error:")) {
            line = line.replace("Error: [unknown] validation error:", "").trim();
            if (!line) return;
        }

        // Extract field path and error message
        const match = line.match(/([^:]+):(.*?)(?:\s+\[[^\]]+\])?$/);
        if (match) {
            let [, fieldPath, errorDetail] = match;
            fieldPath = fieldPath.trim();
            errorDetail = errorDetail.trim();

            // Extract the last part of the field path (e.g., "last_name" from "user.last_name")
            const fieldName = fieldPath.split('.').pop() || fieldPath;

            // Map field name if mapping exists
            const mappedField = fieldMapping[fieldName] || convertSnakeToCamel(fieldName);

            // Use a cleaned error message
            errors[mappedField] = cleanErrorMessage(errorDetail);
        }
    });

    return errors;
}

/**
 * Converts snake_case to camelCase for field names
 */
function convertSnakeToCamel(snakeCase: string): string {
    return snakeCase.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());
}

/**
 * Cleans and formats error messages to be more user-friendly
 */
function cleanErrorMessage(message: string): string {
    // Remove backticks and format regex patterns to be more readable
    message = message.replace(/`([^`]+)`/g, "'$1'");

    // Format common validation errors
    if (message.includes("value does not match regex pattern")) {
        return message.replace("value does not match regex pattern", "contains invalid characters");
    }

    if (message.includes("value length must be")) {
        return message.replace("value length must be", "length must be");
    }

    if (message.includes("value must be")) {
        return message.replace("value must be", "must be");
    }

    return message;
}

/**
 * Unified function to parse all types of validation errors
 */
export function parseErrors(errorMessage: string): FieldErrors {
    // Try parsing as protovalidate error first
    const protoErrors = parseProtoValidateErrors(errorMessage);

    // If we found any proto validation errors, return them
    if (Object.keys(protoErrors).length > 0) {
        return protoErrors;
    }

    // Otherwise, try parsing as legacy validation error
    return parseValidationErrors(errorMessage);
}
