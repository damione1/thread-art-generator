package handlers

import (
	"net/http"

	"github.com/Damione1/thread-art-generator/web/client"
	"github.com/Damione1/thread-art-generator/web/templates"
)

// HomeHandler handles the home page
func HomeHandler(grpcClient *client.GrpcClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Render the home page
		component := templates.Home()
		component.Render(r.Context(), w)
	}
}

// DashboardHandler handles the dashboard page
func DashboardHandler(grpcClient *client.GrpcClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Render the dashboard page
		component := templates.Dashboard()
		component.Render(r.Context(), w)
	}
}
