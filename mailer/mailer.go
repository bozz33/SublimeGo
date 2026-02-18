package mailer

import (
	"fmt"
	"net/smtp"
	"strings"
)

// Mailer is the interface for sending emails.
// Implement it to plug in any email provider (SMTP, SendGrid, Mailgun, etc.).
type Mailer interface {
	Send(msg Message) error
}

// Message represents an outgoing email.
type Message struct {
	To      []string
	Subject string
	Body    string
	HTML    bool
}

// NoopMailer discards all messages (useful for development / testing).
type NoopMailer struct{}

func (n *NoopMailer) Send(msg Message) error {
	return nil
}

// LogMailer prints messages to stdout (useful for development).
type LogMailer struct{}

func (l *LogMailer) Send(msg Message) error {
	fmt.Printf("[mailer] To: %s | Subject: %s\n%s\n", strings.Join(msg.To, ", "), msg.Subject, msg.Body)
	return nil
}

// SMTPConfig holds SMTP connection settings.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// SMTPMailer sends emails via SMTP.
type SMTPMailer struct {
	cfg SMTPConfig
}

// NewSMTPMailer creates a new SMTP mailer.
func NewSMTPMailer(cfg SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (s *SMTPMailer) Send(msg Message) error {
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)

	contentType := "text/plain"
	if msg.HTML {
		contentType = "text/html"
	}

	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: %s; charset=UTF-8\r\n\r\n",
		s.cfg.From,
		strings.Join(msg.To, ", "),
		msg.Subject,
		contentType,
	)

	body := headers + msg.Body
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	return smtp.SendMail(addr, auth, s.cfg.From, msg.To, []byte(body))
}
