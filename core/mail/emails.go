package mail

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

// Emails is the app-facing sender. Mailer stays transport-only.
type Emails struct {
	Mailer  Mailer
	From    Address
	BaseURL string
}

func NewEmails(mailer Mailer, from Address, baseURL string) *Emails {
	return &Emails{
		Mailer:  mailer,
		From:    from,
		BaseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (e *Emails) SendVerifyAccount(ctx context.Context, to Address, token string) error {
	link := e.BaseURL + "/verify?token=" + url.QueryEscape(token)
	name := displayName(to)
	return e.send(ctx, to, "Confirm your ThreadArt account", verifyText(name, link), verifyHTML(name, link))
}

func (e *Emails) SendPasswordReset(ctx context.Context, to Address, token string) error {
	link := e.BaseURL + "/reset-password?token=" + url.QueryEscape(token)
	name := displayName(to)
	return e.send(ctx, to, "Reset your ThreadArt password", resetText(name, link), resetHTML(name, link))
}

func (e *Emails) SendEmailChange(ctx context.Context, to Address, token string) error {
	link := e.BaseURL + "/confirm-email?token=" + url.QueryEscape(token)
	name := displayName(to)
	return e.send(ctx, to, "Confirm your new ThreadArt email", emailChangeText(name, link), emailChangeHTML(name, link))
}

func (e *Emails) send(ctx context.Context, to Address, subject, text, html string) error {
	if e == nil || e.Mailer == nil {
		return fmt.Errorf("mail: emails sender is not configured")
	}
	return e.Mailer.Send(ctx, Message{
		From:    e.From,
		To:      []Address{to},
		Subject: subject,
		Text:    text,
		HTML:    html,
	})
}

func displayName(a Address) string {
	if strings.TrimSpace(a.Name) != "" {
		return strings.TrimSpace(a.Name)
	}
	return a.Email
}

func verifyText(name, link string) string {
	return fmt.Sprintf("Hi %s,\n\nConfirm your ThreadArt account by opening this link (valid 24 hours):\n%s\n\nIf you did not create an account, ignore this email.\n", name, link)
}

func resetText(name, link string) string {
	return fmt.Sprintf("Hi %s,\n\nReset your ThreadArt password by opening this link (valid 1 hour):\n%s\n\nIf you did not request a reset, ignore this email.\n", name, link)
}

func verifyHTML(name, link string) string {
	return renderHTML(htmlParams{
		Preheader: "Confirm your ThreadArt account",
		Heading:   "Confirm your account",
		Body:      fmt.Sprintf("Hi %s, click the button below to confirm your ThreadArt account. This link expires in 24 hours.", name),
		Action:    "Confirm account",
		Link:      link,
		Foot:      "If you did not create an account, you can ignore this email.",
	})
}

func resetHTML(name, link string) string {
	return renderHTML(htmlParams{
		Preheader: "Reset your ThreadArt password",
		Heading:   "Reset your password",
		Body:      fmt.Sprintf("Hi %s, click the button below to choose a new password. This link expires in 1 hour.", name),
		Action:    "Reset password",
		Link:      link,
		Foot:      "If you did not request a reset, you can ignore this email.",
	})
}

func emailChangeText(name, link string) string {
	return fmt.Sprintf("Hi %s,\n\nConfirm your new ThreadArt email by opening this link (valid 24 hours):\n%s\n\nIf you did not request this change, ignore this email.\n", name, link)
}

func emailChangeHTML(name, link string) string {
	return renderHTML(htmlParams{
		Preheader: "Confirm your new ThreadArt email",
		Heading:   "Confirm your new email",
		Body:      fmt.Sprintf("Hi %s, click the button below to confirm this email for your ThreadArt account. This link expires in 24 hours.", name),
		Action:    "Confirm email",
		Link:      link,
		Foot:      "If you did not request this change, you can ignore this email.",
	})
}

type htmlParams struct {
	Preheader string
	Heading   string
	Body      string
	Action    string
	Link      string
	Foot      string
}

var emailHTML = template.Must(template.New("email").Parse(`<!DOCTYPE html>
<html lang="en">
<body style="margin:0;padding:0;background:#0f172a;font-family:ui-sans-serif,system-ui,-apple-system,sans-serif;color:#e2e8f0;">
  <div style="display:none;max-height:0;overflow:hidden;">{{.Preheader}}</div>
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#0f172a;padding:32px 16px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:520px;background:#1e293b;border:1px solid #334155;border-radius:16px;padding:32px;">
          <tr><td style="font-size:20px;font-weight:700;color:#f8fafc;padding-bottom:8px;">ThreadArt</td></tr>
          <tr><td style="font-size:24px;font-weight:700;color:#f8fafc;padding-bottom:16px;">{{.Heading}}</td></tr>
          <tr><td style="font-size:16px;line-height:1.6;color:#cbd5e1;padding-bottom:24px;">{{.Body}}</td></tr>
          <tr>
            <td>
              <a href="{{.Link}}" style="display:inline-block;background:#4f46e5;color:#ffffff;text-decoration:none;font-weight:600;padding:12px 20px;border-radius:10px;">{{.Action}}</a>
            </td>
          </tr>
          <tr><td style="font-size:13px;line-height:1.6;color:#94a3b8;padding-top:24px;">{{.Foot}}</td></tr>
          <tr><td style="font-size:12px;line-height:1.6;color:#64748b;padding-top:16px;word-break:break-all;">{{.Link}}</td></tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`))

func renderHTML(p htmlParams) string {
	var buf bytes.Buffer
	if err := emailHTML.Execute(&buf, p); err != nil {
		return p.Body + "\n" + p.Link
	}
	return buf.String()
}
