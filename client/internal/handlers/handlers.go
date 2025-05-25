package handlers

import (
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	"github.com/Damione1/thread-art-generator/client/internal/templates/pages"
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
	// Get user from context if authenticated (optional for home page)
	user, _ := middleware.UserFromContext(r.Context())

	// Render the home page template
	err := pages.HomePage(user).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// HealthCheckHandler returns a simple 200 OK response for health checks
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
