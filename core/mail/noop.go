package mail

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

// NoopMailer logs and drops. Used when SMTP_HOST is unset (tests, CI).
type NoopMailer struct{}

func (NoopMailer) Send(_ context.Context, msg Message) error {
	to := recipientEmails(msg.To)
	log.Info().
		Strs("to", to).
		Str("subject", msg.Subject).
		Msg("mail: noop (SMTP_HOST unset), dropped")
	return nil
}

// MemoryMailer records messages. Tests only.
type MemoryMailer struct {
	mu       sync.Mutex
	Messages []Message
}

func (m *MemoryMailer) Send(_ context.Context, msg Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, msg)
	return nil
}

func (m *MemoryMailer) Last() (Message, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Messages) == 0 {
		return Message{}, false
	}
	return m.Messages[len(m.Messages)-1], true
}
