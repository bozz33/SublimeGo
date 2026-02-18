package mailer

import (
	"strings"
	"testing"
)

func TestNoopMailer(t *testing.T) {
	m := &NoopMailer{}
	err := m.Send(Message{
		To:      []string{"test@example.com"},
		Subject: "Test",
		Body:    "Hello",
	})
	if err != nil {
		t.Errorf("NoopMailer.Send() returned error: %v", err)
	}
}

func TestLogMailer(t *testing.T) {
	m := &LogMailer{}
	err := m.Send(Message{
		To:      []string{"a@example.com", "b@example.com"},
		Subject: "Hello",
		Body:    "World",
	})
	if err != nil {
		t.Errorf("LogMailer.Send() returned error: %v", err)
	}
}

func TestMessage_MultipleRecipients(t *testing.T) {
	recipients := []string{"a@example.com", "b@example.com", "c@example.com"}
	msg := Message{
		To:      recipients,
		Subject: "Multi",
		Body:    "body",
		HTML:    false,
	}
	if len(msg.To) != 3 {
		t.Errorf("expected 3 recipients, got %d", len(msg.To))
	}
	joined := strings.Join(msg.To, ", ")
	if !strings.Contains(joined, "b@example.com") {
		t.Errorf("expected b@example.com in recipients: %s", joined)
	}
}

func TestSMTPMailer_Config(t *testing.T) {
	cfg := SMTPConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "noreply@example.com",
	}
	m := NewSMTPMailer(cfg)
	if m == nil {
		t.Fatal("NewSMTPMailer returned nil")
	}
	if m.cfg.Host != "smtp.example.com" {
		t.Errorf("expected smtp.example.com, got %s", m.cfg.Host)
	}
	if m.cfg.Port != 587 {
		t.Errorf("expected port 587, got %d", m.cfg.Port)
	}
}

func TestMailer_Interface(t *testing.T) {
	// Verify all implementations satisfy the Mailer interface
	var _ Mailer = &NoopMailer{}
	var _ Mailer = &LogMailer{}
	var _ Mailer = &SMTPMailer{}
}
