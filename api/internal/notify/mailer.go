package notify

// ✮ mailer interface — SMTP for self-hosters, swap for mailjet/SES later

import (
	"context"
	"fmt"

	"github.com/wneessen/go-mail"
)

type Mailer interface {
	Send(ctx context.Context, to, subject, bodyHTML, bodyPlain string) error
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type SMTPMailer struct {
	cfg SMTPConfig
}

func NewSMTPMailer(cfg SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) Send(ctx context.Context, to, subject, bodyHTML, bodyPlain string) error {
	msg := mail.NewMsg()
	if err := msg.From(m.cfg.From); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("mail to: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, bodyHTML)
	if bodyPlain != "" {
		msg.AddAlternativeString(mail.TypeTextPlain, bodyPlain)
	}

	client, err := mail.NewClient(m.cfg.Host,
		mail.WithPort(m.cfg.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(m.cfg.Username),
		mail.WithPassword(m.cfg.Password),
		mail.WithTLSPortPolicy(mail.TLSOpportunistic),
	)
	if err != nil {
		return fmt.Errorf("mail client: %w", err)
	}

	if err := client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("mail send: %w", err)
	}

	return nil
}

// ⋆˙⟡ no-op mailer for dev/testing
type NoopMailer struct{}

func (NoopMailer) Send(_ context.Context, to, subject, _, _ string) error {
	fmt.Printf("mail [noop]: to=%s subject=%s\n", to, subject)
	return nil
}
