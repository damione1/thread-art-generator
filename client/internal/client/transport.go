package client

import (
	"context"
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
)

type incomingCookieKey struct{}

// WithIncomingCookie stores the browser Cookie header so BFF→API RPCs can forward it.
func WithIncomingCookie(ctx context.Context, cookie string) context.Context {
	return context.WithValue(ctx, incomingCookieKey{}, cookie)
}

// IncomingCookieMiddleware copies Cookie onto the request context for the Connect transport.
func IncomingCookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(WithIncomingCookie(r.Context(), r.Header.Get("Cookie"))))
	})
}

// SessionTransport forwards the session cookie to the API. Cookie is the only credential.
type SessionTransport struct {
	SessionManager *auth.SCSSessionManager
	Base           http.RoundTripper
}

func NewSessionTransport(sessionManager *auth.SCSSessionManager) http.RoundTripper {
	return &SessionTransport{
		SessionManager: sessionManager,
		Base:           http.DefaultTransport,
	}
}

func (t *SessionTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if cookie, ok := req.Context().Value(incomingCookieKey{}).(string); ok && cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	return t.Base.RoundTrip(req)
}
