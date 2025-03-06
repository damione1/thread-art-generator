package client

import (
	"net/http"
	"time"

	"github.com/Damione1/thread-art-generator/core/pb"
)

const (
	// SessionCookieName is the name of the cookie that stores the session ID
	SessionCookieName = "thread_art_session"
	// RefreshTokenCookieName is the name of the cookie that stores the refresh token
	RefreshTokenCookieName = "thread_art_refresh_token"
)

// SetSessionCookies sets the session cookies
func SetSessionCookies(w http.ResponseWriter, session *pb.CreateSessionResponse) {
	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    session.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(session.AccessTokenExpireTime.Seconds, 0),
	})

	// Set the refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    session.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(session.RefreshTokenExpireTime.Seconds, 0),
	})
}

// SetRefreshedCookies sets the cookies after a token refresh
func SetRefreshedCookies(w http.ResponseWriter, refreshResp *pb.RefreshTokenResponse) {
	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    refreshResp.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(refreshResp.AccessTokenExpireTime.Seconds, 0),
	})

	// Set the refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    refreshResp.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(refreshResp.RefreshTokenExpireTime.Seconds, 0),
	})
}

// ClearSessionCookies clears the session cookies
func ClearSessionCookies(w http.ResponseWriter) {
	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	// Clear the refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// GetSessionToken gets the session token from the request
func GetSessionToken(r *http.Request) string {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// GetRefreshToken gets the refresh token from the request
func GetRefreshToken(r *http.Request) string {
	cookie, err := r.Cookie(RefreshTokenCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
