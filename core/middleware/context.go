package middleware

import "context"

// AuthKey stores the authenticated user ID on the request context.
const AuthKey = "user_id"

type AdminContext struct {
	context.Context
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(AuthKey).(string)
	return userID, ok
}

func FromAdminContext(ctx context.Context) *AdminContext {
	return ctx.(*AdminContext)
}
