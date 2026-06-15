package notify

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"net/url"
)

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
	if addr.Is4() && addr == netip.MustParseAddr("169.254.169.254") {
		return true
	}
	return false
}
