import { parseValidationErrors } from './errorUtils';

describe('parseValidationErrors', () => {
    it('should parse validation errors correctly', () => {
        const errorMessage = 'failed to validate request: user: (email: cannot be blank; first_name: cannot be blank; last_name: cannot be blank; password: password must contain at least one lowercase letter.)';

        const result = parseValidationErrors(errorMessage);

        expect(result).toEqual({
            email: 'cannot be blank',
            firstName: 'cannot be blank',
            lastName: 'cannot be blank',
            password: 'password must contain at least one lowercase letter'
        });
    });

    it('should handle empty error message', () => {
        const result = parseValidationErrors('');
        expect(result).toEqual({});
    });

    it('should handle non-validation errors', () => {
        const result = parseValidationErrors('Some other error');
        expect(result).toEqual({});
    });

    it('should handle malformed validation errors', () => {
        const result = parseValidationErrors('failed to validate request: user: (malformed error)');
        expect(result).toEqual({});
    });

    it('should handle single field error', () => {
        const result = parseValidationErrors('failed to validate request: user: (email: invalid format)');
        expect(result).toEqual({
            email: 'invalid format'
        });
    });

    it('should handle session validation errors', () => {
        const result = parseValidationErrors('failed to validate request: (email: cannot be blank; password: cannot be blank)');
        expect(result).toEqual({
            email: 'cannot be blank',
            password: 'cannot be blank'
        });
    });

    it('should handle email already exists error', () => {
        const result = parseValidationErrors('failed to validate request: email already exists');
        expect(result).toEqual({
            email: 'Email already exists'
        });
    });

    it('should handle validation number errors', () => {
        const result = parseValidationErrors('failed to validate request: (validation_number: validation number must be 7 digits)');
        expect(result).toEqual({
            validationNumber: 'validation number must be 7 digits'
        });
    });

    it('should handle password complexity errors', () => {
        const result = parseValidationErrors('failed to validate request: user: (password: password must contain at least one uppercase letter)');
        expect(result).toEqual({
            password: 'password must contain at least one uppercase letter'
        });
    });

    it('should handle errors without parentheses', () => {
        const result = parseValidationErrors('failed to validate request: password must be between 10 and 255 characters');
        expect(result).toEqual({
            password: 'password must be between 10 and 255 characters'
        });
    });

    it('should handle invalid resource name errors', () => {
        const result = parseValidationErrors('failed to validate request: invalid resource name: invalid format');
        expect(result).toEqual({
            _generic: 'Invalid resource name'
        });
    });

    it('should handle refresh token errors', () => {
        const result = parseValidationErrors('failed to validate request: (refresh_token: refresh token is required)');
        expect(result).toEqual({
            refreshToken: 'refresh token is required'
        });
    });

    it('should handle authentication errors', () => {
        const result = parseValidationErrors('failed to validate request: (email: incorrect email or password; password: incorrect email or password)');
        expect(result).toEqual({
            email: 'incorrect email or password',
            password: 'incorrect email or password'
        });
    });
});
