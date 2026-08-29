package mail

import (
	"context"
	"strings"
	"testing"

	"github.com/Damione1/thread-art-generator/core/util"
	"github.com/stretchr/testify/require"
)

func TestBuildRFC822Multipart(t *testing.T) {
	t.Parallel()
	raw, err := buildRFC822(Address{Name: "ThreadArt", Email: "noreply@example.com"}, Message{
		To:      []Address{{Name: "Ada", Email: "ada@example.com"}},
		Subject: "Confirm your account",
		Text:    "plain body",
		HTML:    "<p>html body</p>",
	})
	require.NoError(t, err)
	s := string(raw)
	require.Contains(t, s, "From:")
	require.Contains(t, s, "noreply@example.com")
	require.Contains(t, s, "ada@example.com")
	require.Contains(t, s, "multipart/alternative")
	require.Contains(t, s, "plain body")
	require.Contains(t, s, "<p>html body</p>")
	require.Contains(t, s, "MIME-Version: 1.0")
}

func TestBuildRFC822RequiresBodyAndRecipients(t *testing.T) {
	t.Parallel()
	_, err := buildRFC822(Address{Email: "a@b.c"}, Message{To: []Address{{Email: "x@y.z"}}})
	require.Error(t, err)

	_, err = buildRFC822(Address{Email: "a@b.c"}, Message{Text: "hi"})
	require.ErrorIs(t, err, ErrNoRecipients)

	_, err = buildRFC822(Address{}, Message{To: []Address{{Email: "x@y.z"}}, Text: "hi"})
	require.ErrorIs(t, err, ErrNoFrom)
}

func TestNewSMTPMailerValidates(t *testing.T) {
	t.Parallel()
	_, err := NewSMTPMailer(Config{FromAddr: "a@b.c"})
	require.ErrorIs(t, err, ErrNoHost)

	_, err = NewSMTPMailer(Config{Host: "mailhog", TLSMode: "bogus", FromAddr: "a@b.c"})
	require.Error(t, err)

	m, err := NewSMTPMailer(Config{Host: "mailhog", Port: 1025, FromAddr: "a@b.c", TLSMode: "none"})
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestNewMailerNoopWhenHostEmpty(t *testing.T) {
	t.Parallel()
	m, err := NewMailer(Config{})
	require.NoError(t, err)
	require.IsType(t, NoopMailer{}, m)
	require.NoError(t, m.Send(context.Background(), Message{
		To:      []Address{{Email: "a@b.c"}},
		Subject: "x",
		Text:    "y",
	}))
}

func TestConfigFromUtilFallbacks(t *testing.T) {
	t.Parallel()
	got := ConfigFromUtil(util.Config{
		EmailSenderName:     "Legacy",
		EmailSenderAddress:  "legacy@example.com",
		EmailSenderPassword: "legacy-pass",
		SMTP: util.SMTPConfig{
			Host: "mailhog",
			Port: 1025,
		},
	})
	require.Equal(t, "mailhog", got.Host)
	require.Equal(t, 1025, got.Port)
	require.Equal(t, "Legacy", got.FromName)
	require.Equal(t, "legacy@example.com", got.FromAddr)
	require.Equal(t, "legacy-pass", got.Password)
}

func TestEmailsSendVerifyAndReset(t *testing.T) {
	t.Parallel()
	mem := &MemoryMailer{}
	emails := NewEmails(mem, Address{Name: "ThreadArt", Email: "noreply@example.com"}, "http://localhost:8080")

	err := emails.SendVerifyAccount(context.Background(), Address{Name: "Ada", Email: "ada@example.com"}, "tok-verify")
	require.NoError(t, err)
	msg, ok := mem.Last()
	require.True(t, ok)
	require.Equal(t, "Confirm your ThreadArt account", msg.Subject)
	require.Contains(t, msg.Text, "http://localhost:8080/verify?token=tok-verify")
	require.Contains(t, msg.HTML, "Confirm account")
	require.True(t, strings.Contains(msg.HTML, "tok-verify"))

	err = emails.SendPasswordReset(context.Background(), Address{Email: "ada@example.com"}, "tok-reset")
	require.NoError(t, err)
	msg, ok = mem.Last()
	require.True(t, ok)
	require.Equal(t, "Reset your ThreadArt password", msg.Subject)
	require.Contains(t, msg.Text, "http://localhost:8080/reset-password?token=tok-reset")
}
