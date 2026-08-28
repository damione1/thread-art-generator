package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

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
