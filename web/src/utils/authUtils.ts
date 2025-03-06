/**
 * Utility functions for authentication
 */

// Local storage key for the email
const EMAIL_STORAGE_KEY = 'auth_email';

/**
 * Save email to local storage
 * @param email The email to save
 */
export function saveEmail(email: string): void {
    if (typeof window !== 'undefined') {
        localStorage.setItem(EMAIL_STORAGE_KEY, email);
    }
}

/**
 * Get email from local storage
 * @returns The saved email or empty string
 */
export function getSavedEmail(): string {
    if (typeof window !== 'undefined') {
        return localStorage.getItem(EMAIL_STORAGE_KEY) || '';
    }
    return '';
}

/**
 * Clear saved email from local storage
 */
export function clearSavedEmail(): void {
    if (typeof window !== 'undefined') {
        localStorage.removeItem(EMAIL_STORAGE_KEY);
    }
}
