package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
)

const minHMACSecretBytes = 32

var (
	// ErrInvalidServiceSecret is returned when the HMAC secret is too short.
	ErrInvalidServiceSecret = errors.New("service hmac secret must be at least 32 bytes")
	// ErrInvalidServiceID is returned when Sign is given an unusable service id.
	ErrInvalidServiceID = errors.New("invalid service id")
	// ErrInvalidServiceCred is returned when Authorize rejects a header.
	ErrInvalidServiceCred = errors.New("invalid service credential")
)

// HMACServiceAuth authenticates worker/internal calls with HMAC-SHA256.
// Header: Authorization: Service <id>:<hex-mac>
// MAC is over the service id. Timing-safe compare via crypto/subtle.
type HMACServiceAuth struct {
	secret []byte
}

var _ ServiceAuth = (*HMACServiceAuth)(nil)

// NewHMACServiceAuth builds a ServiceAuth from a config secret (32+ bytes).
func NewHMACServiceAuth(secret string) (*HMACServiceAuth, error) {
	if len(secret) < minHMACSecretBytes {
		return nil, ErrInvalidServiceSecret
	}
	return &HMACServiceAuth{secret: []byte(secret)}, nil
}

// Sign returns the full Authorization header value: Service <id>:<hex hmac>.
func (a *HMACServiceAuth) Sign(serviceID string) (string, error) {
	if !validServiceID(serviceID) {
		return "", ErrInvalidServiceID
	}
	return servicePrefix + serviceID + ":" + hex.EncodeToString(a.mac([]byte(serviceID))), nil
}

const servicePrefix = "Service "

// Authorize parses Authorization: Service <id>:<hex-mac> and constant-time
// compares the MAC. ctx is unused (interface).
func (a *HMACServiceAuth) Authorize(_ context.Context, header string) (Identity, error) {
	id, provided, parsed := parseServiceHeader(header)

	expected := a.mac([]byte(id))
	var provided32 [sha256.Size]byte
	lengthOK := 0
	if len(provided) == sha256.Size {
		copy(provided32[:], provided)
		lengthOK = 1
	}

	macOK := subtle.ConstantTimeCompare(provided32[:], expected)
	// Combine flags without short-circuiting the compare itself.
	if subtle.ConstantTimeSelect(parsed&lengthOK&macOK, 1, 0) != 1 {
		return Identity{}, ErrInvalidServiceCred
	}
	return Identity{UserID: id, Kind: PrincipalService}, nil
}

func (a *HMACServiceAuth) mac(msg []byte) []byte {
	m := hmac.New(sha256.New, a.secret)
	_, _ = m.Write(msg)
	return m.Sum(nil)
}

func validServiceID(id string) bool {
	if id == "" {
		return false
	}
	return !strings.ContainsAny(id, " :\t\r\n")
}

// parseServiceHeader returns id, raw MAC bytes, and 1 if the header is well-formed.
// On failure, parsed is 0; id/mac may be dummy so Authorize still runs Compare.
func parseServiceHeader(header string) (id string, mac []byte, parsed int) {
	scheme, rest, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Service") {
		return "", nil, 0
	}
	id, macHex, found := strings.Cut(rest, ":")
	if !found || !validServiceID(id) || macHex == "" {
		return "", nil, 0
	}
	mac, err := hex.DecodeString(macHex)
	if err != nil {
		return id, nil, 0
	}
	return id, mac, 1
}
