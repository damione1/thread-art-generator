package middleware

import (
	"context"
)

// AuthKey is the key used to store the Firebase UID in the context
const AuthKey = "firebase_user_id"

type AdminContext struct {
	context.Context
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

// fromAdminContext retrieves the AdminContext from a regular context.
func FromAdminContext(ctx context.Context) *AdminContext {
	return ctx.(*AdminContext)
}
