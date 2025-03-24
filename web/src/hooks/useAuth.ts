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

    const loginRedirect = useCallback(() => {
        loginWithRedirect({
            appState: {
                returnTo: window.location.pathname,
            },
        });
    }, [loginWithRedirect]);

    const logoutUser = useCallback(() => {
        logout({
            logoutParams: {
                returnTo: window.location.origin,
            },
        });
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
