package handlers

import (
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/templates"
)

// RegisterHandlers registers all HTTP handlers with the provided ServeMux
func RegisterHandlers(mux *http.ServeMux) {
	// Register home page
	mux.HandleFunc("GET /", HomeHandler)

	// Health check endpoint
	mux.HandleFunc("GET /health", HealthCheckHandler)
}

// HomeHandler handles the home page request
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	templates.HomePage(nil).Render(r.Context(), w)
}

// HealthCheckHandler returns a simple 200 OK response for health checks
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
