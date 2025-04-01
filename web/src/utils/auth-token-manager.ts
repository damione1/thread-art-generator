import { getAuth0Client } from "./auth0-client";

/**
 * Get a valid access token from Auth0
 */
export const getAccessToken = async (): Promise<string | undefined> => {
    if (typeof window === "undefined") {
        return undefined; // No token in SSR
    }

    try {
        const auth0 = await getAuth0Client();
        const isAuthenticated = await auth0.isAuthenticated();

        if (!isAuthenticated) {
            return undefined;
        }

        // Always pass audience and scope parameters to ensure consistent tokens
        return auth0.getTokenSilently({
            authorizationParams: {
                audience: process.env.NEXT_PUBLIC_AUTH0_AUDIENCE,
                scope: "openid profile email",
            },
        });
    } catch (error) {
        console.error("Error getting access token:", error);

        // Check if we need to redirect to login
        if (isLoginRequired(error)) {
            redirectToLogin();
        }

        return undefined;
    }
};

/**
 * Determine if the error requires login redirect
 */
function isLoginRequired(error: unknown): boolean {
    if (error && typeof error === "object") {
        // Auth0 error - requires login
        if ("error" in error && error.error === "login_required") {
            return true;
        }

        // Token expired error
        if ("code" in error &&
            (error.code === "invalid_token" ||
                error.code === "expired_token")) {
            return true;
        }
    }
    return false;
}

/**
 * Redirect to login when token refresh fails
 */
function redirectToLogin(): void {
    getAuth0Client().then(client => {
        client.loginWithRedirect({
            appState: {
                returnTo: window.location.pathname,
            },
            authorizationParams: {
                audience: process.env.NEXT_PUBLIC_AUTH0_AUDIENCE,
                scope: "openid profile email",
            }
        });
    }).catch(err => {
        console.error("Failed to redirect to login:", err);
        // Fallback in case auth0 client fails
        window.location.href = "/login";
    });
}

/**
 * Force refresh the token by logging out and reloading the page
 */
export const refreshAccessToken = async (): Promise<string | undefined> => {
    if (typeof window === "undefined") {
        return undefined;
    }

    try {
        const auth0 = await getAuth0Client();
        await auth0.logout({
            openUrl: false,
            logoutParams: {
                returnTo: window.location.origin,
            }
        });
        window.location.reload();
        return undefined;
    } catch (error) {
        console.error("Error refreshing token:", error);
        return undefined;
    }
};
