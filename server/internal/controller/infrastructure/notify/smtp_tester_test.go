package notify

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestSMTPTesterReportsUnavailableSender(t *testing.T) {
	err := NewSMTPTester(nil).SendTestEmail(context.Background(), "admin@example.com")
	if err == nil || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("expected unavailable sender error, got %v", err)
	}
}

func TestSMTPTesterReportsConfigLookupFailure(t *testing.T) {
	provider := &smtpTestConfigProvider{err: errors.New("settings lookup failed")}
	tester := NewSMTPTester(NewDynamicSMTPSender(provider))

	err := tester.SendTestEmail(context.Background(), "admin@example.com")
	if !errors.Is(err, provider.err) {
		t.Fatalf("expected config lookup error, got %v", err)
	}
}

func TestSMTPTesterReportsUnconfiguredSMTP(t *testing.T) {
	tester := NewSMTPTester(NewDynamicSMTPSender(&smtpTestConfigProvider{}))

	err := tester.SendTestEmail(context.Background(), "admin@example.com")
	if err == nil || !strings.Contains(err.Error(), "SMTP is not configured") {
		t.Fatalf("expected unconfigured SMTP error, got %v", err)
	}
}

type smtpTestConfigProvider struct {
	config SMTPConfig
	err    error
}

func (p *smtpTestConfigProvider) SMTPConfig(context.Context) (SMTPConfig, error) {
	return p.config, p.err
}
