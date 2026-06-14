package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"time"

	appnotification "github.com/yorukot/netstamp/internal/controller/application/notification"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

type WebhookSender struct {
	client *http.Client
}

func NewWebhookSender(timeout time.Duration) *WebhookSender {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	sender := &WebhookSender{}
	sender.client = &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			return validateWebhookTarget(req.Context(), req.URL.String())
		},
	}
	return sender
}

func (s *WebhookSender) SendWebhook(ctx context.Context, channel domainalert.NotificationChannel, payload []byte) appnotification.DeliveryResult {
	var config domainalert.WebhookConfig
	if err := json.Unmarshal(channel.Config, &config); err != nil {
		return permanent("config", "invalid_config", "invalid webhook configuration")
	}
	if err := validateWebhookTarget(ctx, config.URL); err != nil {
		return permanent("security", "blocked_target", err.Error())
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.URL, bytes.NewReader(payload))
	if err != nil {
		return permanent("request", "invalid_request", "invalid webhook request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "netstamp-alerts/1")

	resp, err := s.client.Do(req)
	if err != nil {
		return retryable("network", "request_failed", "webhook request failed")
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return appnotification.DeliveryResult{Delivered: true}
	case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusRequestTimeout || resp.StatusCode >= 500:
		return retryable("http", fmt.Sprintf("status_%d", resp.StatusCode), "webhook returned retryable status")
	default:
		return permanent("http", fmt.Sprintf("status_%d", resp.StatusCode), "webhook returned permanent status")
	}
}

func validateWebhookTarget(ctx context.Context, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("invalid webhook URL")
	}
	if parsed.Scheme != "https" {
		return errors.New("webhook URL must use https")
	}
	host := parsed.Hostname()
	if host == "" {
		return errors.New("webhook URL host is required")
	}
	if ip, parseErr := netip.ParseAddr(host); parseErr == nil {
		if blockedAddr(ip) {
			return errors.New("webhook URL points to a blocked address")
		}
		return nil
	}
	addrs, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return errors.New("webhook host could not be resolved")
	}
	if len(addrs) == 0 {
		return errors.New("webhook host resolved no addresses")
	}
	for _, addr := range addrs {
		if blockedAddr(addr) {
			return errors.New("webhook host resolves to a blocked address")
		}
	}
	return nil
}

func blockedAddr(addr netip.Addr) bool {
	if addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsMulticast() || addr.IsUnspecified() {
		return true
	}
	if addr.Is4() {
		if addr == netip.MustParseAddr("169.254.169.254") {
			return true
		}
	}
	return false
}

func retryable(kind, code, message string) appnotification.DeliveryResult {
	return appnotification.DeliveryResult{Retryable: true, Kind: kind, Code: code, Message: message}
}

func permanent(kind, code, message string) appnotification.DeliveryResult {
	return appnotification.DeliveryResult{Retryable: false, Kind: kind, Code: code, Message: message}
}
