/**
 * Utility functions for handling API errors
 */

/**
 * Parse validation errors from the backend
 * The backend returns errors in various formats:
 * 1. "failed to validate request: user: (email: cannot be blank; first_name: cannot be blank; last_name: cannot be blank; password: password must contain at least one lowercase letter.)"
 * 2. "failed to validate request: (email: cannot be blank; password: cannot be blank)"
 * 3. "failed to validate request: email already exists"
 * 4. "failed to validate request: invalid resource name: ..."
 *
 * @param errorMessage The error message from the backend
 * @returns An object with field names as keys and error messages as values
 */
export function parseValidationErrors(errorMessage: string): { [key: string]: string } {
    const errors: { [key: string]: string } = {};

    // Check if it's a validation error
    if (!errorMessage || !errorMessage.includes('failed to validate request')) {
        return errors;
    }

    try {
        // Handle special case for email already exists
        if (errorMessage.includes('email already exists')) {
            errors['email'] = 'Email already exists';
            return errors;
        }

        // Handle special case for invalid resource name
        if (errorMessage.includes('invalid resource name')) {
            errors['_generic'] = 'Invalid resource name';
            return errors;
        }

        // Extract the part inside parentheses
        const match = errorMessage.match(/\((.*?)\)/);
        if (!match || !match[1]) {
            // If no parentheses, check if there's a specific error after the colon
            const colonMatch = errorMessage.match(/failed to validate request:\s*(.*)/);
            if (colonMatch && colonMatch[1]) {
                // This might be a single error like "email already exists"
                const singleError = colonMatch[1].trim();
                if (singleError.includes('email')) {
                    errors['email'] = singleError;
                } else if (singleError.includes('password')) {
                    errors['password'] = singleError;
                } else if (singleError.includes('refresh_token')) {
                    errors['refreshToken'] = singleError;
                } else {
                    // Generic error
                    errors['_generic'] = singleError;
                }
            }
            return errors;
        }

        // Split by semicolon to get individual field errors
        const fieldErrors = match[1].split(';').map(err => err.trim());

        // Parse each field error
        fieldErrors.forEach(fieldError => {
            // Split by colon to get field name and error message
            const parts = fieldError.split(':');
            if (parts.length < 2) return;

            const fieldName = parts[0].trim();
            const errorMsg = parts.slice(1).join(':').trim(); // Join back in case error message contains colons

            // Convert backend field names to frontend field names
            const fieldMapping: { [key: string]: string } = {
                'email': 'email',
                'password': 'password',
                'first_name': 'firstName',
                'last_name': 'lastName',
                'validation_number': 'validationNumber',
                'refresh_token': 'refreshToken'
            };

            const frontendFieldName = fieldMapping[fieldName] || fieldName;
            if (frontendFieldName && errorMsg) {
                errors[frontendFieldName] = errorMsg;
            }
        });
    } catch (error) {
        console.error('Error parsing validation errors:', error);
    }

    return errors;
}
