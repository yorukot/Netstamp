package notify

import (
	"context"
	"errors"
	"fmt"
)

const (
	smtpTestSubject = "Netstamp SMTP test"
	smtpTestBody    = "Your Netstamp SMTP configuration is working."
)

type SMTPTester struct {
	sender *DynamicSMTPSender
}

func NewSMTPTester(sender *DynamicSMTPSender) *SMTPTester {
	return &SMTPTester{sender: sender}
}

func (t *SMTPTester) SendTestEmail(ctx context.Context, recipient string) error {
	if t == nil || t.sender == nil {
		return errors.New("SMTP test sender is unavailable")
	}
	sender, err := t.sender.Sender(ctx)
	if err != nil {
		return err
	}
	result := sender.SendMessage(ctx, EmailMessage{
		To:      []string{recipient},
		Subject: smtpTestSubject,
		Body:    smtpTestBody,
	})
	if !result.Delivered {
		return fmt.Errorf("SMTP test delivery failed: %s", result.Message)
	}
	return nil
}
