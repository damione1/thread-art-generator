package mail

import (
	"strings"

	"github.com/Damione1/thread-art-generator/core/util"
)

// NewMailer returns SMTP when Host is set, otherwise NoopMailer.
func NewMailer(cfg Config) (Mailer, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return NoopMailer{}, nil
	}
	return NewSMTPMailer(cfg)
}

func ConfigFromUtil(c util.Config) Config {
	fromName := c.SMTP.FromName
	if fromName == "" {
		fromName = c.EmailSenderName
	}
	fromAddr := c.SMTP.FromAddr
	if fromAddr == "" {
		fromAddr = c.EmailSenderAddress
	}
	password := c.SMTP.Password
	if password == "" {
		password = c.EmailSenderPassword
	}
	return Config{
		Host:     c.SMTP.Host,
		Port:     c.SMTP.Port,
		Username: c.SMTP.Username,
		Password: password,
		FromName: fromName,
		FromAddr: fromAddr,
		TLSMode:  c.SMTP.TLSMode,
	}
}
