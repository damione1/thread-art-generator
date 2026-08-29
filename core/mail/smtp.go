package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type smtpMailer struct {
	cfg Config
}

// NewSMTPMailer builds a Mailer. Does not dial; first Send is the health check.
func NewSMTPMailer(cfg Config) (Mailer, error) {
	cfg.TLSMode = normalizeTLSMode(cfg.TLSMode)
	if err := validateSMTPConfig(cfg); err != nil {
		return nil, err
	}
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	return &smtpMailer{cfg: cfg}, nil
}

func validateSMTPConfig(cfg Config) error {
	if strings.TrimSpace(cfg.Host) == "" {
		return ErrNoHost
	}
	if strings.TrimSpace(cfg.FromAddr) == "" {
		return ErrNoFrom
	}
	switch normalizeTLSMode(cfg.TLSMode) {
	case TLSNone, TLSStartTLS, TLSImplicit:
	default:
		return fmt.Errorf("mail: unknown smtp tls mode %q (none|starttls|tls)", cfg.TLSMode)
	}
	return nil
}

func (s *smtpMailer) Send(ctx context.Context, msg Message) error {
	from := msg.From
	if strings.TrimSpace(from.Email) == "" {
		from = s.cfg.fromAddress()
	}
	raw, err := buildRFC822(from, msg)
	if err != nil {
		return err
	}
	recipients := recipientEmails(msg.To)
	if len(recipients) == 0 {
		return ErrNoRecipients
	}

	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))
	client, err := s.dial(ctx, addr)
	if err != nil {
		return fmt.Errorf("mail: smtp dial: %w", err)
	}
	defer client.Close()

	if err := s.handshake(client); err != nil {
		return err
	}
	if err := client.Mail(from.Email); err != nil {
		return fmt.Errorf("mail: MAIL FROM: %w", err)
	}
	for _, rcpt := range recipients {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("mail: RCPT TO %s: %w", rcpt, err)
		}
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("mail: DATA: %w", err)
	}
	if _, err := wc.Write(raw); err != nil {
		_ = wc.Close()
		return fmt.Errorf("mail: write body: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("mail: close body: %w", err)
	}
	return client.Quit()
}

func (s *smtpMailer) dial(ctx context.Context, addr string) (*smtp.Client, error) {
	d := net.Dialer{Timeout: 10 * time.Second}
	if deadline, ok := ctx.Deadline(); ok {
		d.Deadline = deadline
	}

	if s.cfg.TLSMode == TLSImplicit {
		tlsCfg := &tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12}
		dialer := &tls.Dialer{NetDialer: &d, Config: tlsCfg}
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			return nil, err
		}
		return smtp.NewClient(conn, s.cfg.Host)
	}

	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return smtp.NewClient(conn, s.cfg.Host)
}

func (s *smtpMailer) handshake(client *smtp.Client) error {
	if s.cfg.TLSMode == TLSStartTLS {
		tlsCfg := &tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12}
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("mail: STARTTLS: %w", err)
		}
	}

	if strings.TrimSpace(s.cfg.Username) == "" {
		return nil
	}
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("mail: AUTH: %w", err)
	}
	return nil
}
