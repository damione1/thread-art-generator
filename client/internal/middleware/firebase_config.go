package middleware

import (
	"context"
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/types"
	"github.com/Damione1/thread-art-generator/core/util"
)

// firebaseConfigKey is used to store Firebase config in request context
type firebaseConfigKey struct{}

// FirebaseConfigMiddleware adds Firebase configuration to the request context for authenticated users
func FirebaseConfigMiddleware(config *util.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if user is authenticated (user will be in context from FirebaseAuthMiddleware)
			user, ok := UserFromContext(r.Context())
			if ok && user != nil {
				// Get Firebase configuration
				coreConfig := config.GetFirebaseConfigForFrontend()
				
				// Convert to client types format
				firebaseConfig := &types.FirebaseConfig{
					ProjectID:    coreConfig.ProjectID,
					APIKey:       coreConfig.APIKey,
					AuthDomain:   coreConfig.AuthDomain,
					EmulatorHost: coreConfig.EmulatorHost,
					EmulatorUI:   coreConfig.EmulatorUI,
					IsEmulator:   coreConfig.IsEmulator,
				}
				
				// Add Firebase config to request context
				ctx := context.WithValue(r.Context(), firebaseConfigKey{}, firebaseConfig)
				r = r.WithContext(ctx)
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// FirebaseConfigFromContext retrieves Firebase config from request context
func FirebaseConfigFromContext(ctx context.Context) (*types.FirebaseConfig, bool) {
	config, ok := ctx.Value(firebaseConfigKey{}).(*types.FirebaseConfig)
	return config, ok
}