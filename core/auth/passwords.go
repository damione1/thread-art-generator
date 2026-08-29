package auth

import (
	"errors"

	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/alexedwards/argon2id"
)

// ErrMismatchedPassword is returned by Passwords.Check when the password does not match.
var ErrMismatchedPassword = errors.New("mismatched password")

// Argon2idPasswords implements Passwords via alexedwards/argon2id (OWASP current default).
type Argon2idPasswords struct{}

var _ Passwords = Argon2idPasswords{}

func (Argon2idPasswords) Hash(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func (Argon2idPasswords) Check(password, hash string) error {
	ok, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return err
	}
	if !ok {
		return ErrMismatchedPassword
	}
	return nil
}

// BcryptPasswords implements Passwords via core/util (bcrypt). Kept for tests / legacy hashes.
type BcryptPasswords struct{}

var _ Passwords = BcryptPasswords{}

func (BcryptPasswords) Hash(password string) (string, error) {
	return util.HashPassword(password)
}

func (BcryptPasswords) Check(password, hash string) error {
	return util.CheckPassword(password, hash)
}
