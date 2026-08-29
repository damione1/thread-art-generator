package middleware

import (
	"context"
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
)

type csrfContextKey struct{}

const (
	csrfHeader    = "X-CSRF-Token"
	csrfFormField = "csrf_token"
)

func CSRFFromContext(ctx context.Context) string {
	token, _ := ctx.Value(csrfContextKey{}).(string)
	return token
}

func withCSRF(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, csrfContextKey{}, token)
}

// CSRFMiddleware issues a per-session token and requires it (+ Origin/Referer)
// on mutating requests. Safe methods only mint the token.
func CSRFMiddleware(sessionManager *auth.SCSSessionManager, allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := sessionManager.EnsureCSRFToken(r)
			if err != nil {
				http.Error(w, "failed to issue csrf token", http.StatusInternalServerError)
				return
			}
			r = r.WithContext(withCSRF(r.Context(), token))

			if csrfSafeMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			if !originAllowed(r, allowedOrigin) {
				http.Error(w, "forbidden origin", http.StatusForbidden)
				return
			}
			if !validCSRF(token, submittedCSRF(r)) {
				http.Error(w, "invalid csrf token", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func csrfSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func submittedCSRF(r *http.Request) string {
	if t := strings.TrimSpace(r.Header.Get(csrfHeader)); t != "" {
		return t
	}
	_ = r.ParseForm()
	return strings.TrimSpace(r.FormValue(csrfFormField))
}

func validCSRF(expected, got string) bool {
	if expected == "" || got == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(got)) == 1
}

func originAllowed(r *http.Request, allowedOrigin string) bool {
	hosts := allowedHosts(allowedOrigin, r.Host)
	if raw := r.Header.Get("Origin"); raw != "" && raw != "null" {
		return hostIn(raw, hosts)
	}
	if raw := r.Header.Get("Referer"); raw != "" {
		return hostIn(raw, hosts)
	}
	return false
}

func allowedHosts(allowedOrigin, reqHost string) map[string]struct{} {
	out := make(map[string]struct{}, 2)
	if u, err := url.Parse(allowedOrigin); err == nil && u.Host != "" {
		out[strings.ToLower(u.Host)] = struct{}{}
	}
	if h := strings.ToLower(strings.TrimSpace(reqHost)); h != "" {
		out[h] = struct{}{}
	}
	return out
}

func hostIn(raw string, hosts map[string]struct{}) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	_, ok := hosts[strings.ToLower(u.Host)]
	return ok
}

