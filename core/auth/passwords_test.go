package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestValidatePasswordLength(t *testing.T) {
	t.Parallel()
	require.ErrorIs(t, ValidatePasswordLength("short"), ErrPasswordTooShort)
	require.NoError(t, ValidatePasswordLength("password12"))
	long := strings.Repeat("a", MaxPasswordLength+1)
	require.ErrorIs(t, ValidatePasswordLength(long), ErrPasswordTooLong)
	_, err := Argon2idPasswords{}.Hash(long)
	require.ErrorIs(t, err, ErrPasswordTooLong)
}

func TestArgon2idPasswords(t *testing.T) {
	p := Argon2idPasswords{}
	hash, err := p.Hash("hunter2-secret")
	require.NoError(t, err)
	require.NoError(t, p.Check("hunter2-secret", hash))
	require.ErrorIs(t, p.Check("wrong", hash), ErrMismatchedPassword)
}

func TestBcryptPasswords(t *testing.T) {
	p := BcryptPasswords{}
	hash, err := p.Hash("hunter2-secret")
	require.NoError(t, err)
	require.NoError(t, p.Check("hunter2-secret", hash))
	require.ErrorIs(t, p.Check("wrong", hash), bcrypt.ErrMismatchedHashAndPassword)
}
