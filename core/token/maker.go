package token

import "time"

type Maker interface {
	// CreateToken creates a token for the given username and duration.
	CreateToken(UserID string, duration time.Duration) (string, *Payload, error)

	// ValidateToken check if the token is valid or not.
	ValidateToken(token string) (*Payload, error)
}
