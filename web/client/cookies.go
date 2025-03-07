package client

import (
	"net/http"
)

const (
	// EmailCookieName is the name of the cookie that stores the user's email
	EmailCookieName = "user_email"
	// EmailCookieMaxAge is the max age of the email cookie in seconds (30 days)
	EmailCookieMaxAge = 60 * 60 * 24 * 30
)

// SetEmailCookie stores the user's email in a cookie
func SetEmailCookie(w http.ResponseWriter, email string) {
	cookie := &http.Cookie{
		Name:     EmailCookieName,
		Value:    email,
		Path:     "/",
		MaxAge:   EmailCookieMaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// GetEmailFromCookie retrieves the user's email from the cookie
func GetEmailFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(EmailCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// ClearEmailCookie removes the email cookie
func ClearEmailCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     EmailCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}
