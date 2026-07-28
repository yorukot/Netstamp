package config

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"
)

func validateRequiredString(key, value string) []error {
	if strings.TrimSpace(value) == "" {
		return []error{fmt.Errorf("%s must not be empty", key)}
	}
	return nil
}

func validateLogLevel(value string) []error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return []error{errors.New("LOG_LEVEL must not be empty")}
	}
	if _, err := zapcore.ParseLevel(strings.ToLower(trimmed)); err != nil {
		return []error{errors.New("LOG_LEVEL must be one of debug, info, warn, error, dpanic, panic, or fatal")}
	}
	return nil
}

func validateAPIVersion(value string) []error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return []error{errors.New("API_VERSION must not be empty")}
	}
	if !strings.HasPrefix(trimmed, "v") {
		return []error{errors.New("API_VERSION must start with 'v'")}
	}
	if strings.ContainsAny(trimmed, "/?#") || strings.Contains(trimmed, "..") {
		return []error{errors.New("API_VERSION must be a single URL path segment")}
	}
	return nil
}

func validateListenAddr(key, value string) []error {
	if strings.TrimSpace(value) == "" {
		return []error{fmt.Errorf("%s must not be empty", key)}
	}

	_, port, err := net.SplitHostPort(value)
	if err != nil {
		return []error{fmt.Errorf("%s must be a host:port address", key)}
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return []error{fmt.Errorf("%s port must be between 1 and 65535", key)}
	}

	return nil
}

func validatePositiveDuration(key string, value time.Duration) []error {
	if value <= 0 {
		return []error{fmt.Errorf("%s must be greater than 0", key)}
	}
	return nil
}

func validatePositiveUint32(key string, value uint32) []error {
	if value == 0 {
		return []error{fmt.Errorf("%s must be greater than 0", key)}
	}
	return nil
}

func validatePositiveUint8(key string, value uint8) []error {
	if value == 0 {
		return []error{fmt.Errorf("%s must be greater than 0", key)}
	}
	return nil
}

func validateDatabasePort(value int32) []error {
	if value < 1 || value > 65535 {
		return []error{errors.New("DATABASE_PORT must be between 1 and 65535")}
	}
	return nil
}

func validateDatabaseSSLMode(value string) []error {
	switch strings.TrimSpace(value) {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
		return nil
	default:
		return []error{errors.New("DATABASE_SSLMODE must be one of disable, allow, prefer, require, verify-ca, or verify-full")}
	}
}

func validateOptionalHTTPURL(key, value string) []error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	parsed, err := url.ParseRequestURI(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return []error{fmt.Errorf("%s must be a valid HTTP URL", key)}
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return []error{fmt.Errorf("%s must use http or https", key)}
	}

	return nil
}

func validateOptionalHTTPOrigin(key, value string) []error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" {
		return []error{fmt.Errorf("%s must be a valid HTTP origin", key)}
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return []error{fmt.Errorf("%s must use http or https", key)}
	}
	if parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return []error{fmt.Errorf("%s must be an origin without path, query, fragment, or credentials", key)}
	}

	return nil
}

func validateTrustedProxyPrefixes(key, value string) []error {
	if _, err := parseTrustedProxyPrefixes(value); err != nil {
		return []error{fmt.Errorf("%s %w", key, err)}
	}

	return nil
}

func parseTrustedProxyPrefixes(value string) ([]netip.Prefix, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	parts := strings.Split(trimmed, ",")
	prefixes := make([]netip.Prefix, 0, len(parts))
	for _, part := range parts {
		raw := strings.TrimSpace(part)
		if raw == "" {
			return nil, errors.New("must not contain empty entries")
		}

		prefix, err := parseTrustedProxyPrefix(raw)
		if err != nil {
			return nil, fmt.Errorf("must contain valid IP addresses or CIDR prefixes: %w", err)
		}
		prefixes = append(prefixes, prefix)
	}

	return prefixes, nil
}

func parseTrustedProxyPrefix(value string) (netip.Prefix, error) {
	if prefix, err := netip.ParsePrefix(value); err == nil {
		return prefix.Masked(), nil
	}

	addr, err := netip.ParseAddr(value)
	if err != nil {
		return netip.Prefix{}, err
	}
	if addr.Is4() {
		return netip.PrefixFrom(addr, 32), nil
	}

	return netip.PrefixFrom(addr, 128), nil
}
