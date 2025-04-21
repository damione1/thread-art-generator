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
        error,
    } = useAuth0();

    const getToken = useCallback(async () => {
        try {
            // Always include audience to ensure consistency
            return await getAccessTokenSilently({
                authorizationParams: {
                    audience: process.env.NEXT_PUBLIC_AUTH0_AUDIENCE,
                    scope: 'openid profile email',
                }
            });
        } catch (error) {
            console.error('Error getting token:', error);

            // If login required, redirect
            if (error && typeof error === 'object' && 'error' in error && error.error === 'login_required') {
                loginRedirect();
            }

            return undefined;
        }
    }, [getAccessTokenSilently, loginWithRedirect]);

    const loginRedirect = useCallback((options?: Record<string, unknown>) => {
        try {
            loginWithRedirect(options || {
                appState: {
                    returnTo: window.location.pathname,
                },
                authorizationParams: {
                    audience: process.env.NEXT_PUBLIC_AUTH0_AUDIENCE,
                    scope: 'openid profile email',
                }
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
        error,
    };
}
