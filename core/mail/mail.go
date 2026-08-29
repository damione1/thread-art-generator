package mail

import (
	"context"
	"errors"
	"strings"
)

// TLS modes for SMTP. "none" is Mailhog / unencrypted relay.
const (
	TLSNone     = "none"
	TLSStartTLS = "starttls"
	TLSImplicit = "tls"
)

var (
	ErrNoRecipients = errors.New("mail: no recipients")
	ErrNoFrom       = errors.New("mail: missing from address")
	ErrNoHost       = errors.New("mail: missing smtp host")
)

// Address is a named mailbox.
type Address struct {
	Name  string
	Email string
}

// Message is a single outbound email. HTML and Text are both optional; at
// least one must be set. From is optional and falls back to Config.
type Message struct {
	From    Address
	To      []Address
	Subject string
	Text    string
	HTML    string
}

// Mailer sends email. One implementation: SMTP. Tests use Memory.
type Mailer interface {
	Send(ctx context.Context, msg Message) error
}

// Config is SMTP transport. Secrets live in env (SMTP_USERNAME / SMTP_PASSWORD).
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	FromName string
	FromAddr string
	TLSMode  string // none | starttls | tls
}

func (c Config) fromAddress() Address {
	return Address{Name: c.FromName, Email: c.FromAddr}
}

func normalizeTLSMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", TLSNone:
		return TLSNone
	case TLSStartTLS, "start_tls":
		return TLSStartTLS
	case TLSImplicit, "smtps":
		return TLSImplicit
	default:
		return strings.ToLower(strings.TrimSpace(mode))
	}
}
