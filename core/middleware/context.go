package middleware

import (
	"context"

	"github.com/Damione1/thread-art-generator/core/token"
)

// AuthKey is the key used to store the Firebase UID in the context
const AuthKey = "firebase_user_id"

type AdminContext struct {
	context.Context
	UserPayload *token.Payload
}

// UserIDFromContext retrieves the Firebase UID from the context
func UserIDFromContext(ctx context.Context) (string, bool) {
	firebaseUID, ok := ctx.Value(AuthKey).(string)
	return firebaseUID, ok
}

// FirebaseUIDFromContext is an alias for UserIDFromContext for clarity
func FirebaseUIDFromContext(ctx context.Context) (string, bool) {
	return UserIDFromContext(ctx)
}

// newAdminContext creates a new AdminContext from a regular context and user payload.
func NewAdminContext(ctx context.Context, userPayload *token.Payload) *AdminContext {
	return &AdminContext{
		Context:     ctx,
		UserPayload: userPayload,
	}
}

// fromAdminContext retrieves the AdminContext from a regular context.
func FromAdminContext(ctx context.Context) *AdminContext {
	return ctx.(*AdminContext)
}
