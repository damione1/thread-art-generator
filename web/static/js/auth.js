/**
 * Auth utilities for the frontend
 */

// Token storage keys
const ACCESS_TOKEN_KEY = "thread_art_access_token";
const REFRESH_TOKEN_KEY = "thread_art_refresh_token";
const TOKEN_EXPIRY_KEY = "thread_art_token_expiry";

/**
 * Store authentication tokens in localStorage
 * @param {string} accessToken - The JWT access token
 * @param {string} refreshToken - The refresh token
 * @param {number} expiryTimestamp - Expiry timestamp in seconds
 */
function storeTokens(accessToken, refreshToken, expiryTimestamp) {
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
  localStorage.setItem(TOKEN_EXPIRY_KEY, expiryTimestamp);
}

/**
 * Get the stored access token
 * @returns {string|null} The access token or null if not found
 */
function getAccessToken() {
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

/**
 * Get the stored refresh token
 * @returns {string|null} The refresh token or null if not found
 */
function getRefreshToken() {
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

/**
 * Check if the access token is expired
 * @returns {boolean} True if token is expired or doesn't exist
 */
function isTokenExpired() {
  const expiry = localStorage.getItem(TOKEN_EXPIRY_KEY);
  if (!expiry) return true;

  // Add a 30-second buffer to account for clock differences
  return parseInt(expiry, 10) - 30 < Math.floor(Date.now() / 1000);
}

/**
 * Clear all auth tokens from localStorage
 */
function clearTokens() {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
  localStorage.removeItem(TOKEN_EXPIRY_KEY);
}

/**
 * Refresh the access token using the refresh token
 * @returns {Promise<boolean>} Whether the refresh was successful
 */
async function refreshAccessToken() {
  const refreshToken = getRefreshToken();
  if (!refreshToken) return false;

  try {
    const response = await fetch("/api/auth/refresh", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!response.ok) {
      clearTokens();
      return false;
    }

    const data = await response.json();
    storeTokens(
      data.access_token,
      data.refresh_token,
      data.access_token_expire_time
    );
    return true;
  } catch (error) {
    console.error("Failed to refresh token:", error);
    clearTokens();
    return false;
  }
}

/**
 * Make an authenticated fetch request with automatic token refresh
 * @param {string} url - The URL to fetch
 * @param {Object} options - Fetch options
 * @returns {Promise<Response>} The fetch response
 */
async function authFetch(url, options = {}) {
  // Check if token is expired and refresh if needed
  if (isTokenExpired()) {
    const refreshed = await refreshAccessToken();
    if (!refreshed) {
      // Redirect to login if refresh failed
      window.location.href = "/login";
      throw new Error("Authentication failed");
    }
  }

  // Get the current access token
  const token = getAccessToken();

  // Set up headers with authorization
  const headers = options.headers || {};
  headers["Authorization"] = `Bearer ${token}`;

  // Make the authenticated request
  const response = await fetch(url, {
    ...options,
    headers,
  });

  // If we get a 401, try to refresh the token and retry once
  if (response.status === 401) {
    const refreshed = await refreshAccessToken();
    if (refreshed) {
      // Retry the request with the new token
      headers["Authorization"] = `Bearer ${getAccessToken()}`;
      return fetch(url, {
        ...options,
        headers,
      });
    } else {
      // Redirect to login if refresh failed
      window.location.href = "/login";
      throw new Error("Authentication failed");
    }
  }

  return response;
}

/**
 * Submit a form with authentication
 * @param {HTMLFormElement} form - The form to submit
 * @returns {Promise<Response>} The fetch response
 */
async function submitAuthForm(form) {
  const formData = new FormData(form);
  const url = form.action;
  const method = form.method.toUpperCase();

  return authFetch(url, {
    method,
    body: formData,
  });
}

// Export the utilities
window.authUtils = {
  storeTokens,
  getAccessToken,
  getRefreshToken,
  isTokenExpired,
  clearTokens,
  refreshAccessToken,
  authFetch,
  submitAuthForm,
};
