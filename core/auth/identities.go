package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
)

// ErrIdentityNotFound is returned by Identities.ByEmail when the email is unknown.
var ErrIdentityNotFound = errors.New("identity not found")

// PGIdentities looks up users by email using raw SQL so password_hash can
// land without regenerating sqlboiler models.
type PGIdentities struct {
	DB *sql.DB
}

var _ Identities = (*PGIdentities)(nil)

func (p *PGIdentities) ByEmail(ctx context.Context, email string) (Identity, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return Identity{}, "", ErrIdentityNotFound
	}
	var (
		id, hash, first string
		last            sql.NullString
	)
	err := p.DB.QueryRowContext(ctx, `
		SELECT id, COALESCE(password_hash, ''), first_name, last_name
		FROM users
		WHERE lower(email) = $1
	`, email).Scan(&id, &hash, &first, &last)
	if errors.Is(err, sql.ErrNoRows) {
		return Identity{}, "", ErrIdentityNotFound
	}
	if err != nil {
		return Identity{}, "", err
	}
	return Identity{
		UserID: id,
		Email:  email,
		Kind:   PrincipalUser,
	}, hash, nil
}

func (p *PGIdentities) Create(ctx context.Context, email, passwordHash, first, last string) (Identity, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return Identity{}, errors.New("email is required")
	}
	if first == "" {
		first = "User"
	}
	id := uuid.New().String()
	_, err := p.DB.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, active, role)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), true, 'user')
	`, id, email, passwordHash, first, last)
	if err != nil {
		return Identity{}, err
	}
	return Identity{
		UserID: id,
		Email:  email,
		Kind:   PrincipalUser,
	}, nil
}
