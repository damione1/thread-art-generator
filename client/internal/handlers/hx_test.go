package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedirectHTMX(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/dashboard/arts/new", nil)
	req.Header.Set("HX-Request", "true")
	redirect(rec, req, "/dashboard/arts/abc")
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "/dashboard/arts/abc", rec.Header().Get("HX-Redirect"))
}

func TestRedirectBrowser(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/dashboard/arts/new", nil)
	redirect(rec, req, "/dashboard/arts/abc")
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Equal(t, "/dashboard/arts/abc", rec.Header().Get("Location"))
}
