package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/stretchr/testify/require"
)

func TestCSRFAllowsGETWithoutToken(t *testing.T) {
	sm := auth.NewInMemorySessionManager()
	h := sm.GetSessionManager().LoadAndSave(CSRFMiddleware(sm, "http://localhost:8080")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NotEmpty(t, CSRFFromContext(r.Context()))
		w.WriteHeader(http.StatusOK)
	})))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/login", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotEmpty(t, rec.Result().Cookies())
}

func TestCSRFRejectsPOSTWithoutToken(t *testing.T) {
	sm := auth.NewInMemorySessionManager()
	h := csrfTestHandler(sm)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	cookie := rec.Result().Cookies()[0]

	post := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{}`))
	post.Header.Set("Content-Type", "application/json")
	post.Header.Set("Origin", "http://localhost:8080")
	post.AddCookie(cookie)
	postRec := httptest.NewRecorder()
	h.ServeHTTP(postRec, post)
	require.Equal(t, http.StatusForbidden, postRec.Code)
}

func TestCSRFAcceptsMatchingTokenAndOrigin(t *testing.T) {
	sm := auth.NewInMemorySessionManager()
	var token string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token = CSRFFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	h := sm.GetSessionManager().LoadAndSave(CSRFMiddleware(sm, "http://localhost:8080")(inner))

	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/", nil))
	cookie := getRec.Result().Cookies()[0]
	require.NotEmpty(t, token)

	post := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{}`))
	post.Header.Set("Content-Type", "application/json")
	post.Header.Set("Origin", "http://localhost:8080")
	post.Header.Set("X-CSRF-Token", token)
	post.AddCookie(cookie)
	postRec := httptest.NewRecorder()
	h.ServeHTTP(postRec, post)
	require.Equal(t, http.StatusOK, postRec.Code)
}

func TestCSRFRejectsForeignOrigin(t *testing.T) {
	sm := auth.NewInMemorySessionManager()
	var token string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token = CSRFFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	h := sm.GetSessionManager().LoadAndSave(CSRFMiddleware(sm, "http://localhost:8080")(inner))

	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/", nil))
	cookie := getRec.Result().Cookies()[0]

	post := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	post.Header.Set("Origin", "https://evil.example")
	post.Header.Set("X-CSRF-Token", token)
	post.AddCookie(cookie)
	postRec := httptest.NewRecorder()
	h.ServeHTTP(postRec, post)
	require.Equal(t, http.StatusForbidden, postRec.Code)
}

func TestCSRFFormField(t *testing.T) {
	sm := auth.NewInMemorySessionManager()
	var token string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token = CSRFFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	h := sm.GetSessionManager().LoadAndSave(CSRFMiddleware(sm, "http://localhost:8080")(inner))

	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/", nil))
	cookie := getRec.Result().Cookies()[0]

	post := httptest.NewRequest(http.MethodPost, "/settings/profile", strings.NewReader("csrf_token="+token+"&first_name=Ada"))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	post.Header.Set("Origin", "http://localhost:8080")
	post.AddCookie(cookie)
	postRec := httptest.NewRecorder()
	h.ServeHTTP(postRec, post)
	require.Equal(t, http.StatusOK, postRec.Code)
}

func TestCSRFFormFieldWithoutOrigin(t *testing.T) {
	sm := auth.NewInMemorySessionManager()
	var token string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token = CSRFFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	h := sm.GetSessionManager().LoadAndSave(CSRFMiddleware(sm, "http://localhost:8080")(inner))

	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/", nil))
	cookie := getRec.Result().Cookies()[0]

	post := httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader("csrf_token="+token))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	post.AddCookie(cookie)
	postRec := httptest.NewRecorder()
	h.ServeHTTP(postRec, post)
	require.Equal(t, http.StatusOK, postRec.Code)
}

func TestCSRFRejectsCrossSiteFetchSite(t *testing.T) {
	sm := auth.NewInMemorySessionManager()
	var token string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token = CSRFFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	h := sm.GetSessionManager().LoadAndSave(CSRFMiddleware(sm, "http://localhost:8080")(inner))

	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/", nil))
	cookie := getRec.Result().Cookies()[0]

	post := httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader("csrf_token="+token))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	post.Header.Set("Sec-Fetch-Site", "cross-site")
	post.AddCookie(cookie)
	postRec := httptest.NewRecorder()
	h.ServeHTTP(postRec, post)
	require.Equal(t, http.StatusForbidden, postRec.Code)
}

func csrfTestHandler(sm *auth.SCSSessionManager) http.Handler {
	return sm.GetSessionManager().LoadAndSave(CSRFMiddleware(sm, "http://localhost:8080")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
}
