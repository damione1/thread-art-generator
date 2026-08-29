package auth

import (
	"errors"
	"sync"
	"unicode/utf8"

	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/alexedwards/argon2id"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 128
)

// ErrMismatchedPassword is returned by Passwords.Check when the password does not match.
var ErrMismatchedPassword = errors.New("mismatched password")

// ErrPasswordTooShort is returned when a password is below MinPasswordLength.
var ErrPasswordTooShort = errors.New("password must be at least 8 characters")

// ErrPasswordTooLong is returned when a password exceeds MaxPasswordLength.
var ErrPasswordTooLong = errors.New("password is too long")

// ValidatePasswordLength enforces min/max before hashing. Length is runes, not bytes.
func ValidatePasswordLength(password string) error {
	n := utf8.RuneCountInString(password)
	if n < MinPasswordLength {
		return ErrPasswordTooShort
	}
	if n > MaxPasswordLength {
		return ErrPasswordTooLong
	}
	return nil
}

// Argon2idPasswords implements Passwords via alexedwards/argon2id (OWASP current default).
type Argon2idPasswords struct{}

var _ Passwords = Argon2idPasswords{}

func (Argon2idPasswords) Hash(password string) (string, error) {
	if err := ValidatePasswordLength(password); err != nil {
		return "", err
	}
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

var (
	dummyHashOnce sync.Once
	dummyHash     string
)

func dummyArgon2Hash() string {
	dummyHashOnce.Do(func() {
		h, err := argon2id.CreateHash("timing-dummy-password-not-used", argon2id.DefaultParams)
		if err != nil {
			return
		}
		dummyHash = h
	})
	return dummyHash
}

// CheckDummy runs an argon2 compare against a unused hash so unknown-email
// logins take roughly as long as known-email misses.
func (p Argon2idPasswords) CheckDummy(password string) {
	h := dummyArgon2Hash()
	if h == "" {
		return
	}
	_, _ = argon2id.ComparePasswordAndHash(password, h)
}

// BcryptPasswords implements Passwords via core/util (bcrypt). Kept for tests / legacy hashes.
type BcryptPasswords struct{}

var _ Passwords = BcryptPasswords{}

func (BcryptPasswords) Hash(password string) (string, error) {
	if err := ValidatePasswordLength(password); err != nil {
		return "", err
	}
	return util.HashPassword(password)
}

func (BcryptPasswords) Check(password, hash string) error {
	return util.CheckPassword(password, hash)
}
