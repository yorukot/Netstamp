package notify

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	appnotification "github.com/yorukot/netstamp/internal/controller/application/notification"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

const (
	emailSubjectLimit = 180
	emailBodyLimit    = 12000
)

type SMTPConfig struct {
	Host     string
	Port     int32
	Username string
	Password string
	From     string
	TLSMode  string
	Timeout  time.Duration
}

type SMTPSender struct {
	cfg SMTPConfig
	now func() time.Time
}

type EmailMessage struct {
	To      []string
	Subject string
	Body    string
}

type AlertEmailSender struct {
	smtp *SMTPSender
}

type PasswordResetMailer struct {
	smtp *SMTPSender
}

func NewSMTPSender(cfg SMTPConfig) *SMTPSender {
	cfg.Host = strings.TrimSpace(cfg.Host)
	cfg.Username = strings.TrimSpace(cfg.Username)
	cfg.From = strings.TrimSpace(cfg.From)
	cfg.TLSMode = strings.TrimSpace(cfg.TLSMode)
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	if cfg.TLSMode == "" {
		cfg.TLSMode = "starttls"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &SMTPSender{cfg: cfg, now: func() time.Time { return time.Now().UTC() }}
}

func (s *SMTPSender) Configured() bool {
	return strings.TrimSpace(s.cfg.Host) != "" && strings.TrimSpace(s.cfg.From) != ""
}

func NewAlertEmailSender(smtp *SMTPSender) *AlertEmailSender {
	return &AlertEmailSender{smtp: smtp}
}

func (s *AlertEmailSender) Configured() bool {
	return s.smtp != nil && s.smtp.Configured()
}

func (s *AlertEmailSender) Send(ctx context.Context, notification domainalert.Notification, payload []byte) appnotification.DeliveryResult {
	if s.smtp == nil || !s.smtp.Configured() {
		return permanent("config", "smtp_unconfigured", "SMTP notifications are not configured")
	}

	_, config, err := domainalert.VNEmailConfig(notification.Config)
	if err != nil {
		return permanent("config", "invalid_config", "invalid email configuration")
	}
	return s.smtp.SendAlert(ctx, config, payload)
}

func NewPasswordResetMailer(cfg SMTPConfig) *PasswordResetMailer {
	return &PasswordResetMailer{smtp: NewSMTPSender(cfg)}
}

func (m *PasswordResetMailer) SendPasswordReset(ctx context.Context, input identity.PasswordResetEmail) error {
	if m == nil || m.smtp == nil || !m.smtp.Configured() {
		return errors.New("SMTP is not configured")
	}

	result := m.smtp.SendMessage(ctx, EmailMessage{
		To:      []string{input.To},
		Subject: "Reset your Netstamp password",
		Body:    renderPasswordResetBody(input),
	})
	if deliveryFailed(result) {
		return errors.New(result.Message)
	}

	return nil
}

func (s *SMTPSender) SendAlert(ctx context.Context, config domainalert.EmailConfig, payload []byte) appnotification.DeliveryResult {
	message, renderErr := renderEmailMessage(payload, s.cfg.From, config.To, s.now())
	if renderErr != nil {
		return permanent("request", "invalid_request", "invalid email request")
	}

	return s.sendRaw(ctx, config.To, message)
}

func (s *SMTPSender) SendMessage(ctx context.Context, message EmailMessage) appnotification.DeliveryResult {
	if !s.Configured() {
		return permanent("config", "smtp_unconfigured", "SMTP is not configured")
	}
	if len(message.To) == 0 || strings.TrimSpace(message.Subject) == "" || strings.TrimSpace(message.Body) == "" {
		return permanent("request", "invalid_request", "invalid email request")
	}

	rawMessage, renderErr := renderPlainEmailMessage(message, s.cfg.From, s.now())
	if renderErr != nil {
		return permanent("request", "invalid_request", "invalid email request")
	}

	return s.sendRaw(ctx, message.To, rawMessage)
}

func (s *SMTPSender) sendRaw(ctx context.Context, recipients []string, message []byte) appnotification.DeliveryResult {
	client, err := s.newClient(ctx)
	if err != nil {
		return retryable("network", "connect_failed", "SMTP connection failed")
	}
	defer client.Close()

	if result := s.configureSecurity(client); deliveryFailed(result) {
		return result
	}
	if result := s.authenticate(client); deliveryFailed(result) {
		return result
	}
	if mailErr := client.Mail(s.cfg.From); mailErr != nil {
		return smtpDeliveryResult(mailErr, "mail_from_failed", "SMTP sender was rejected")
	}
	for _, recipient := range recipients {
		if recipientErr := client.Rcpt(recipient); recipientErr != nil {
			return smtpDeliveryResult(recipientErr, "recipient_rejected", "SMTP recipient was rejected")
		}
	}
	writer, err := client.Data()
	if err != nil {
		return smtpDeliveryResult(err, "data_failed", "SMTP DATA command failed")
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return retryable("network", "write_failed", "SMTP message write failed")
	}
	if err := writer.Close(); err != nil {
		return smtpDeliveryResult(err, "message_rejected", "SMTP message was rejected")
	}
	if err := client.Quit(); err != nil {
		return retryable("network", "quit_failed", "SMTP quit failed")
	}
	return appnotification.DeliveryResult{Delivered: true}
}

func (s *SMTPSender) newClient(ctx context.Context) (*smtp.Client, error) {
	address := net.JoinHostPort(s.cfg.Host, strconv.FormatInt(int64(s.cfg.Port), 10))
	dialer := net.Dialer{Timeout: s.cfg.Timeout}

	var conn net.Conn
	var err error
	if s.cfg.TLSMode == "implicit" {
		tlsDialer := tls.Dialer{NetDialer: &dialer, Config: s.tlsConfig()}
		conn, err = tlsDialer.DialContext(ctx, "tcp", address)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", address)
	}
	if err != nil {
		return nil, err
	}
	if deadlineErr := conn.SetDeadline(time.Now().Add(s.cfg.Timeout)); deadlineErr != nil {
		_ = conn.Close()
		return nil, deadlineErr
	}

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return client, nil
}

func (s *SMTPSender) configureSecurity(client *smtp.Client) appnotification.DeliveryResult {
	switch s.cfg.TLSMode {
	case "implicit", "none":
		return appnotification.DeliveryResult{Delivered: true}
	case "starttls":
		ok, _ := client.Extension("STARTTLS")
		if !ok {
			return permanent("smtp", "starttls_unavailable", "SMTP server does not support STARTTLS")
		}
		if err := client.StartTLS(s.tlsConfig()); err != nil {
			return smtpDeliveryResult(err, "starttls_failed", "SMTP STARTTLS failed")
		}
		return appnotification.DeliveryResult{Delivered: true}
	default:
		return permanent("config", "invalid_tls_mode", "SMTP TLS mode is invalid")
	}
}

func (s *SMTPSender) authenticate(client *smtp.Client) appnotification.DeliveryResult {
	if s.cfg.Username == "" && s.cfg.Password == "" {
		return appnotification.DeliveryResult{Delivered: true}
	}
	if s.cfg.Username == "" || s.cfg.Password == "" {
		return permanent("config", "invalid_auth", "SMTP username and password must be set together")
	}
	if s.cfg.TLSMode == "none" {
		return permanent("config", "auth_requires_tls", "SMTP authentication requires TLS")
	}
	if err := client.Auth(smtpPlainAuth{username: s.cfg.Username, password: s.cfg.Password}); err != nil {
		return smtpDeliveryResult(err, "auth_failed", "SMTP authentication failed")
	}
	return appnotification.DeliveryResult{Delivered: true}
}

func (s *SMTPSender) tlsConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: s.cfg.Host,
	}
}

func renderEmailMessage(payload []byte, from string, to []string, at time.Time) ([]byte, error) {
	subject, body := renderEmailContent(payload)
	return renderPlainEmailMessage(EmailMessage{To: to, Subject: subject, Body: body}, from, at)
}

func renderPlainEmailMessage(input EmailMessage, from string, at time.Time) ([]byte, error) {
	headers := []string{
		"From: " + formatEmailSender(from),
		"To: " + formatEmailRecipients(input.To),
		"Subject: " + mime.QEncoding.Encode("utf-8", truncateMessage(input.Subject, emailSubjectLimit)),
		"Date: " + at.UTC().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
	}

	message := strings.Join(headers, "\r\n") + "\r\n\r\n" + truncateMessage(input.Body, emailBodyLimit) + "\r\n"
	return []byte(message), nil
}

func renderEmailContent(payload []byte) (string, string) {
	incident, ok := parseIncidentNotificationPayload(payload)
	if !ok {
		return "Netstamp alert", "Netstamp alert\n\n" + string(payload)
	}

	subject := incidentNotificationTitle(incident)
	if incident.EventType != "notification.test" && incident.Rule.Name != "" {
		subject += ": " + incident.Rule.Name
	}

	lines := []string{incidentNotificationTitle(incident)}
	if description := incidentNotificationDescription(incident); description != "" {
		lines = append(lines, "", description)
	}
	for _, field := range incidentNotificationFields(incident) {
		lines = append(lines, field.Name+": "+field.Value)
	}
	return subject, strings.Join(lines, "\n")
}

func formatEmailRecipients(values []string) string {
	recipients := make([]string, 0, len(values))
	for _, value := range values {
		address := mail.Address{Address: value}
		recipients = append(recipients, address.String())
	}
	return strings.Join(recipients, ", ")
}

func formatEmailSender(from string) string {
	address := mail.Address{Name: "Netstamp", Address: from}
	return address.String()
}

func renderPasswordResetBody(input identity.PasswordResetEmail) string {
	name := strings.TrimSpace(input.DisplayName)
	if name == "" {
		name = "Netstamp user"
	}

	lines := []string{
		"Hello " + name + ",",
		"",
		"We received a request to reset the password for your Netstamp account.",
		"",
		input.ResetURL,
		"",
		"This link expires at " + input.ExpiresAt.UTC().Format(time.RFC3339) + ".",
		"If you did not request this change, you can ignore this email.",
	}

	return strings.Join(lines, "\n")
}

func smtpDeliveryResult(err error, fallbackCode, message string) appnotification.DeliveryResult {
	var smtpErr *textproto.Error
	if errors.As(err, &smtpErr) {
		code := fmt.Sprintf("smtp_%d", smtpErr.Code)
		if smtpErr.Code >= 400 && smtpErr.Code < 500 {
			return retryable("smtp", code, message)
		}
		return permanent("smtp", code, message)
	}
	return retryable("smtp", fallbackCode, message)
}

func deliveryFailed(result appnotification.DeliveryResult) bool {
	return !result.Delivered && (result.Retryable || result.Kind != "" || result.Code != "" || result.Message != "")
}

type smtpPlainAuth struct {
	username string
	password string
}

func (a smtpPlainAuth) Start(*smtp.ServerInfo) (string, []byte, error) {
	return "PLAIN", []byte("\x00" + a.username + "\x00" + a.password), nil
}

func (a smtpPlainAuth) Next([]byte, bool) ([]byte, error) {
	return nil, nil
}
