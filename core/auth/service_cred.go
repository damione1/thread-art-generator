package auth

import "context"

// ServiceAuth authenticates worker/internal gRPC calls.
// Header: Authorization: Service <id>:<hex-mac>
// Timing-safe. Not a user token. Not a presign.
type ServiceAuth interface {
	Authorize(ctx context.Context, header string) (Identity, error)
	Sign(serviceID string) (string, error)
}
