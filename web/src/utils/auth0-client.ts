import { Auth0Client } from '@auth0/auth0-spa-js';
import { Auth0ClientOptions } from '@auth0/auth0-spa-js';

let auth0Client: Auth0Client | null = null;

/**
 * Get the Auth0 client instance, creating it if it doesn't exist
 */
export const getAuth0Client = async (): Promise<Auth0Client> => {
    if (auth0Client) {
        return auth0Client;
    }

    const options: Auth0ClientOptions = {
        domain: process.env.NEXT_PUBLIC_AUTH0_DOMAIN || '',
        clientId: process.env.NEXT_PUBLIC_AUTH0_CLIENT_ID || '',
        authorizationParams: {
            audience: process.env.NEXT_PUBLIC_AUTH0_AUDIENCE,
            redirect_uri: typeof window !== 'undefined' ? window.location.origin + '/dashboard' : '',
            scope: 'openid profile email read:current_user update:current_user_metadata',
        },
        cacheLocation: 'localstorage',
    };

    auth0Client = new Auth0Client(options);

    // Try to handle the callback if we're on the callback page
    if (
        typeof window !== 'undefined' &&
        window.location.search.includes('code=') &&
        window.location.search.includes('state=')
    ) {
        try {
            await auth0Client.handleRedirectCallback();
        } catch (error) {
            console.error('Error handling redirect callback:', error);
        }
    }

    return auth0Client;
};
