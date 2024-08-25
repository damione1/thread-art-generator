package middleware

import (
	"context"

	"github.com/Damione1/thread-art-generator/core/token"
)

type AdminContext struct {
	context.Context
	UserPayload *token.Payload
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
