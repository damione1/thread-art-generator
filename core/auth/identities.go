package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ErrIdentityNotFound is returned by Identities.ByEmail when the email is unknown.
var ErrIdentityNotFound = errors.New("identity not found")

// ErrEmailTaken is returned by Create/UpdateProfile on unique email collision.
var ErrEmailTaken = errors.New("email already exists")

// ErrAccountInactive is returned by login when the account is not verified.
var ErrAccountInactive = errors.New("account is not verified")

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
	return p.scanAccount(ctx, `
		SELECT id, COALESCE(email, ''), COALESCE(password_hash, ''), first_name, last_name, active
		FROM users
		WHERE lower(email) = $1
	`, email)
}

func (p *PGIdentities) ByID(ctx context.Context, userID string) (Identity, string, error) {
	if strings.TrimSpace(userID) == "" {
		return Identity{}, "", ErrIdentityNotFound
	}
	return p.scanAccount(ctx, `
		SELECT id, COALESCE(email, ''), COALESCE(password_hash, ''), first_name, last_name, active
		FROM users
		WHERE id = $1
	`, userID)
}

func (p *PGIdentities) scanAccount(ctx context.Context, query string, arg any) (Identity, string, error) {
	var (
		id, email, hash, first string
		last                   sql.NullString
		active                 bool
	)
	err := p.DB.QueryRowContext(ctx, query, arg).Scan(&id, &email, &hash, &first, &last, &active)
	if errors.Is(err, sql.ErrNoRows) {
		return Identity{}, "", ErrIdentityNotFound
	}
	if err != nil {
		return Identity{}, "", err
	}
	return Identity{
		UserID:    id,
		Email:     email,
		FirstName: first,
		LastName:  last.String,
		Active:    active,
		Kind:      PrincipalUser,
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
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), false, 'user')
	`, id, email, passwordHash, first, last)
	if isUniqueViolation(err) {
		return Identity{}, ErrEmailTaken
	}
	if err != nil {
		return Identity{}, err
	}
	return Identity{
		UserID:    id,
		Email:     email,
		FirstName: first,
		LastName:  last,
		Active:    false,
		Kind:      PrincipalUser,
	}, nil
}

func (p *PGIdentities) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	res, err := p.DB.ExecContext(ctx, `
		UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1
	`, userID, passwordHash)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrIdentityNotFound
	}
	return nil
}

func (p *PGIdentities) UpdateProfile(ctx context.Context, userID, first, last, email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(first) == "" {
		return errors.New("first name is required")
	}
	res, err := p.DB.ExecContext(ctx, `
		UPDATE users
		SET first_name = $2, last_name = NULLIF($3, ''), email = $4, updated_at = NOW()
		WHERE id = $1
	`, userID, first, last, email)
	if isUniqueViolation(err) {
		return ErrEmailTaken
	}
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrIdentityNotFound
	}
	return nil
}

func (p *PGIdentities) SetActive(ctx context.Context, userID string, active bool) error {
	res, err := p.DB.ExecContext(ctx, `
		UPDATE users SET active = $2, updated_at = NOW() WHERE id = $1
	`, userID, active)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrIdentityNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
