package token

import (
	"fmt"
	"time"

	"golang.org/x/crypto/chacha20poly1305"

	"github.com/o1egl/paseto"
)

type PasetoMaker struct {
	paseto *paseto.V2
	symKey []byte
}

func NewPasetoMaker(symKey string) (*PasetoMaker, error) {
	if len(symKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("symmetric key must be exactly %d characters long", chacha20poly1305.KeySize)
	}

	maker := &PasetoMaker{
		paseto: paseto.NewV2(),
		symKey: []byte(symKey),
	}

	return maker, nil

}

func (maker *PasetoMaker) CreateToken(userID string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(userID, duration)
	if err != nil {
		return "", payload, err
	}

	token, err := maker.paseto.Encrypt(maker.symKey, payload, nil)
	return token, payload, err
}

func (maker *PasetoMaker) ValidateToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := maker.paseto.Decrypt(token, maker.symKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if err := payload.Valid(); err != nil {
		return nil, err
	}

	return payload, nil
}
