package mail

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/mail"
	"strings"
	"time"
)

func buildRFC822(from Address, msg Message) ([]byte, error) {
	if strings.TrimSpace(from.Email) == "" {
		return nil, ErrNoFrom
	}
	if len(msg.To) == 0 {
		return nil, ErrNoRecipients
	}
	if strings.TrimSpace(msg.Text) == "" && strings.TrimSpace(msg.HTML) == "" {
		return nil, fmt.Errorf("mail: message has no body")
	}

	var buf bytes.Buffer
	writeHeader(&buf, "From", formatAddress(from))
	writeHeader(&buf, "To", joinAddresses(msg.To))
	writeHeader(&buf, "Subject", encodeHeader(msg.Subject))
	writeHeader(&buf, "Date", time.Now().UTC().Format(time.RFC1123Z))
	writeHeader(&buf, "MIME-Version", "1.0")
	writeHeader(&buf, "Message-ID", generateMessageID(from.Email))

	text := msg.Text
	html := msg.HTML
	switch {
	case text != "" && html != "":
		boundary := randomBoundary()
		writeHeader(&buf, "Content-Type", `multipart/alternative; boundary="`+boundary+`"`)
		buf.WriteString("\r\n")
		writePart(&buf, boundary, "text/plain; charset=UTF-8", text)
		writePart(&buf, boundary, "text/html; charset=UTF-8", html)
		fmt.Fprintf(&buf, "--%s--\r\n", boundary)
	case html != "":
		writeHeader(&buf, "Content-Type", "text/html; charset=UTF-8")
		buf.WriteString("\r\n")
		buf.WriteString(html)
		if !strings.HasSuffix(html, "\r\n") {
			buf.WriteString("\r\n")
		}
	default:
		writeHeader(&buf, "Content-Type", "text/plain; charset=UTF-8")
		buf.WriteString("\r\n")
		buf.WriteString(text)
		if !strings.HasSuffix(text, "\r\n") {
			buf.WriteString("\r\n")
		}
	}

	return buf.Bytes(), nil
}

func writeHeader(w io.Writer, key, value string) {
	fmt.Fprintf(w, "%s: %s\r\n", key, value)
}

func writePart(buf *bytes.Buffer, boundary, contentType, body string) {
	fmt.Fprintf(buf, "--%s\r\n", boundary)
	writeHeader(buf, "Content-Type", contentType)
	writeHeader(buf, "Content-Transfer-Encoding", "8bit")
	buf.WriteString("\r\n")
	buf.WriteString(body)
	if !strings.HasSuffix(body, "\r\n") {
		buf.WriteString("\r\n")
	}
}

func formatAddress(a Address) string {
	email := strings.TrimSpace(a.Email)
	name := strings.TrimSpace(a.Name)
	if name == "" {
		return email
	}
	addr := mail.Address{Name: name, Address: email}
	return addr.String()
}

func joinAddresses(addrs []Address) string {
	parts := make([]string, 0, len(addrs))
	for _, a := range addrs {
		if strings.TrimSpace(a.Email) == "" {
			continue
		}
		parts = append(parts, formatAddress(a))
	}
	return strings.Join(parts, ", ")
}

func encodeHeader(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	return mime.QEncoding.Encode("utf-8", s)
}

func randomBoundary() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("b%x", time.Now().UnixNano())
	}
	return "b" + hex.EncodeToString(b[:])
}

func generateMessageID(fromEmail string) string {
	domain := "localhost"
	if i := strings.LastIndex(fromEmail, "@"); i >= 0 && i+1 < len(fromEmail) {
		domain = fromEmail[i+1:]
	}
	var b [8]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("<%s.%d@%s>", hex.EncodeToString(b[:]), time.Now().UnixNano(), domain)
}

func recipientEmails(to []Address) []string {
	out := make([]string, 0, len(to))
	for _, a := range to {
		email := strings.TrimSpace(a.Email)
		if email != "" {
			out = append(out, email)
		}
	}
	return out
}
