package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecurityHeaders(t *testing.T) {
	t.Parallel()
	h := SecurityHeaders("http://localhost:8080", "http://localhost:9000/thread-art", true)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	require.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	require.Equal(t, "no-referrer", rec.Header().Get("Referrer-Policy"))
	require.Contains(t, rec.Header().Get("Strict-Transport-Security"), "max-age=")
	csp := rec.Header().Get("Content-Security-Policy")
	require.Contains(t, csp, "frame-ancestors 'none'")
	require.Contains(t, csp, "http://localhost:9000")
	require.Contains(t, csp, "script-src 'self' 'unsafe-eval'")
}

func TestOriginOf(t *testing.T) {
	t.Parallel()
	require.Equal(t, "http://localhost:9000", originOf("http://localhost:9000/thread-art"))
	require.Empty(t, originOf(""))
}
