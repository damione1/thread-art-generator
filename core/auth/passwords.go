package auth

import "github.com/Damione1/thread-art-generator/core/util"

// BcryptPasswords implements Passwords via core/util (bcrypt).
type BcryptPasswords struct{}

var _ Passwords = BcryptPasswords{}

func (BcryptPasswords) Hash(password string) (string, error) {
	return util.HashPassword(password)
}

func (BcryptPasswords) Check(password, hash string) error {
	return util.CheckPassword(password, hash)
}
