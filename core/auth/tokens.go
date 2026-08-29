package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Damione1/thread-art-generator/core/clock"
	"github.com/google/uuid"
)

const (
	TokenVerify TokenPurpose = "verify"
	TokenReset  TokenPurpose = "reset"

	VerifyTTL = 24 * time.Hour
	ResetTTL  = time.Hour
)

// TokenPurpose is stored on email_tokens.purpose.
type TokenPurpose string

// ErrTokenInvalid covers unknown, expired, and already-used tokens.
var ErrTokenInvalid = errors.New("invalid or expired token")

// Tokens issues and consumes hashed email-action tokens.
type Tokens interface {
	Issue(ctx context.Context, userID string, purpose TokenPurpose, ttl time.Duration) (string, error)
	Consume(ctx context.Context, raw string, purpose TokenPurpose) (string, error)
}

// PGTokens stores SHA-256(token) in Postgres. Raw token is returned once.
type PGTokens struct {
	DB    *sql.DB
	Clock clock.Clock
}

var _ Tokens = (*PGTokens)(nil)

func (p *PGTokens) now() time.Time {
	if p.Clock != nil {
		return p.Clock.Now()
	}
	return time.Now()
}

func (p *PGTokens) Issue(ctx context.Context, userID string, purpose TokenPurpose, ttl time.Duration) (string, error) {
	if userID == "" {
		return "", errors.New("user id is required")
	}
	raw, err := newRawToken()
	if err != nil {
		return "", err
	}
	now := p.now()
	_, err = p.DB.ExecContext(ctx, `
		UPDATE email_tokens SET used_at = $3
		WHERE user_id = $1 AND purpose = $2 AND used_at IS NULL
	`, userID, string(purpose), now)
	if err != nil {
		return "", fmt.Errorf("invalidate prior tokens: %w", err)
	}
	_, err = p.DB.ExecContext(ctx, `
		INSERT INTO email_tokens (id, user_id, purpose, token_hash, expiration, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, uuid.New().String(), userID, string(purpose), HashToken(raw), now.Add(ttl), now)
	if err != nil {
		return "", fmt.Errorf("insert token: %w", err)
	}
	return raw, nil
}

func (p *PGTokens) Consume(ctx context.Context, raw string, purpose TokenPurpose) (string, error) {
	if raw == "" {
		return "", ErrTokenInvalid
	}
	now := p.now()
	var userID string
	err := p.DB.QueryRowContext(ctx, `
		UPDATE email_tokens
		SET used_at = $3
		WHERE token_hash = $1 AND purpose = $2 AND used_at IS NULL AND expiration > $3
		RETURNING user_id
	`, HashToken(raw), string(purpose), now).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrTokenInvalid
	}
	if err != nil {
		return "", err
	}
	return userID, nil
}

func newRawToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

// HashToken is SHA-256 hex. Stored value, never the raw token.
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
