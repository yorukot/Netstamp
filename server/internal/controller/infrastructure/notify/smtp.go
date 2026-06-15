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

func (s *WebhookSender) SendEmail(ctx context.Context, notification domainalert.Notification, payload []byte) appnotification.DeliveryResult {
	if s.smtp == nil || !s.smtp.Configured() {
		return permanent("config", "smtp_unconfigured", "SMTP notifications are not configured")
	}

	_, config, err := domainalert.VNEmailConfig(notification.Config)
	if err != nil {
		return permanent("config", "invalid_config", "invalid email configuration")
	}
	return s.smtp.Send(ctx, config, payload)
}

func (s *SMTPSender) Send(ctx context.Context, config domainalert.EmailConfig, payload []byte) appnotification.DeliveryResult {
	message, err := renderEmailMessage(payload, s.cfg.From, config.To, s.now())
	if err != nil {
		return permanent("request", "invalid_request", "invalid email request")
	}

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
	if err := client.Mail(s.cfg.From); err != nil {
		return smtpDeliveryResult(err, "mail_from_failed", "SMTP sender was rejected")
	}
	for _, recipient := range config.To {
		if err := client.Rcpt(recipient); err != nil {
			return smtpDeliveryResult(err, "recipient_rejected", "SMTP recipient was rejected")
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
		conn, err = tls.DialWithDialer(&dialer, "tcp", address, s.tlsConfig())
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", address)
	}
	if err != nil {
		return nil, err
	}
	if err := conn.SetDeadline(time.Now().Add(s.cfg.Timeout)); err != nil {
		_ = conn.Close()
		return nil, err
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
	headers := []string{
		"From: " + formatEmailSender(from),
		"To: " + formatEmailRecipients(to),
		"Subject: " + mime.QEncoding.Encode("utf-8", truncateMessage(subject, emailSubjectLimit)),
		"Date: " + at.UTC().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
	}

	message := strings.Join(headers, "\r\n") + "\r\n\r\n" + truncateMessage(body, emailBodyLimit) + "\r\n"
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
