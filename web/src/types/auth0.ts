/**
 * Auth0User interface for the user object returned by Auth0
 */
export interface Auth0User {
    sub: string;            // Auth0 user ID
    name?: string;          // User's full name
    given_name?: string;    // First name
    family_name?: string;   // Last name
    nickname?: string;      // Nickname
    email?: string;         // Email address
    email_verified?: boolean; // Whether email is verified
    picture?: string;       // Profile picture URL
    locale?: string;        // Locale
    updated_at?: string;    // Last update timestamp
    [key: string]: string | boolean | undefined; // Allow other string or boolean properties
}
