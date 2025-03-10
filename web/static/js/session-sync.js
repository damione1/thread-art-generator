/**
 * This script extracts session tokens from cookies and stores them in localStorage
 * This is needed because our backend sets httpOnly cookies, but our frontend needs
 * to access tokens for API calls.
 */
document.addEventListener("DOMContentLoaded", function () {
  // Check if auth utils are available
  if (!window.authUtils) {
    console.error("Auth utilities not loaded");
    return;
  }

  // Function to extract token data from the page
  function extractTokenData() {
    const tokenDataElement = document.getElementById("token-data");
    if (!tokenDataElement) return null;

    try {
      return JSON.parse(tokenDataElement.textContent);
    } catch (e) {
      console.error("Failed to parse token data:", e);
      return null;
    }
  }

  // Extract tokens and store them
  const tokenData = extractTokenData();
  if (tokenData) {
    window.authUtils.storeTokens(
      tokenData.accessToken,
      tokenData.refreshToken,
      tokenData.expiryTimestamp
    );
    console.log("Session tokens synchronized to localStorage");
  }
});
