import { useAuth0 } from '@auth0/auth0-react';
import { useCallback } from 'react';

export function useAuth() {
    const {
        isAuthenticated,
        isLoading,
        user,
        getAccessTokenSilently,
        loginWithRedirect,
        logout,
    } = useAuth0();

    const getToken = useCallback(async () => {
        try {
            return await getAccessTokenSilently();
        } catch (error) {
            console.error('Error getting token:', error);
            return undefined;
        }
    }, [getAccessTokenSilently]);

    const loginRedirect = useCallback((options?: Record<string, unknown>) => {
        try {
            console.log("Redirecting to Auth0 login...");
            loginWithRedirect(options || {
                appState: {
                    returnTo: window.location.pathname,
                },
            });
        } catch (error) {
            console.error('Failed to redirect to login:', error);
        }
    }, [loginWithRedirect]);

    const logoutUser = useCallback(() => {
        try {
            logout({
                logoutParams: {
                    returnTo: window.location.origin,
                },
            });
        } catch (error) {
            console.error('Failed to logout:', error);
            window.location.href = '/';
        }
    }, [logout]);

    return {
        isAuthenticated,
        isLoading,
        user,
        getToken,
        loginRedirect,
        logoutUser,
    };
}
