package systemsettings

import (
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
)

var dnsLabelPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

func invalidField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}

func validateSMTP(settings SMTPSettings) error {
	var collector appvalidation.Collector
	if settings.Port < 1 || settings.Port > 65535 {
		collector.Add("port", "port must be between 1 and 65535", settings.Port)
	}
	if settings.TimeoutSeconds <= 0 {
		collector.Add("timeoutSeconds", "timeoutSeconds must be greater than 0", settings.TimeoutSeconds)
	}
	switch settings.TLSMode {
	case "starttls", "implicit", "none":
	default:
		collector.Add("tlsMode", "tlsMode must be starttls, implicit, or none", settings.TLSMode)
	}
	if settings.TLSMode == "none" && settings.Username != "" && settings.PasswordSet {
		collector.Add("tlsMode", "tlsMode must use TLS when SMTP authentication is configured", settings.TLSMode)
	}
	if smtpPartiallyConfigured(settings) {
		if strings.TrimSpace(settings.Host) == "" {
			collector.Add("host", "host is required when SMTP is configured", settings.Host)
		} else if err := validateSMTPHost(settings.Host); err != nil {
			collector.Add("host", err.Error(), settings.Host)
		}
		if err := validateSMTPFrom(settings.From); err != nil {
			collector.Add("from", err.Error(), settings.From)
		}
		if (strings.TrimSpace(settings.Username) == "") != !settings.PasswordSet {
			collector.Add("username", "username and password must be set together", settings.Username)
		}
	}
	return collector.Err(ErrInvalidInput)
}

func smtpPartiallyConfigured(settings SMTPSettings) bool {
	return strings.TrimSpace(settings.Host) != "" ||
		strings.TrimSpace(settings.Username) != "" ||
		settings.PasswordSet ||
		strings.TrimSpace(settings.From) != ""
}

func smtpDeliveryConfigured(settings SMTPSettings) bool {
	return strings.TrimSpace(settings.Host) != "" && strings.TrimSpace(settings.From) != ""
}

func validateSMTPFrom(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("from is required when SMTP is configured")
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address == "" {
		return errors.New("from must be a valid email address")
	}
	return nil
}

func validateSMTPHost(value string) error {
	host := strings.TrimSpace(value)
	if net.ParseIP(host) != nil {
		return nil
	}
	if err := validateDNSName(strings.ToLower(host)); err != nil {
		return errors.New("host must be a DNS name or IP address without a scheme or port")
	}
	return nil
}

func validateAccessSMTPInvariant(access AccessSettings, smtp SMTPSettings) error {
	if !access.EmailVerificationRequired {
		return nil
	}
	if validationErr := validateSMTP(smtp); validationErr != nil {
		return validationErr
	}
	if !smtpDeliveryConfigured(smtp) {
		return invalidField(
			"emailVerificationRequired",
			"email verification requires configured SMTP host and from address",
			true,
		)
	}
	return nil
}

func validateOIDC(settings OIDCSettings) error {
	if !settings.Enabled {
		return nil
	}
	var collector appvalidation.Collector
	if settings.IssuerURL == "" {
		collector.Add("issuerUrl", "issuerUrl is required when enabled", settings.IssuerURL)
	} else if err := validateIssuerURL(settings.IssuerURL); err != nil {
		collector.Add("issuerUrl", err.Error(), settings.IssuerURL)
	}
	validateProviderCommon(
		&collector,
		settings.ClientID,
		settings.ClientSecretSet,
		settings.DisplayName,
		settings.CallbackURL,
	)
	return collector.Err(ErrInvalidInput)
}

func validateGoogle(settings GoogleSettings) error {
	var collector appvalidation.Collector
	if _, err := normalizeGoogleDomains(settings.AllowedDomains); err != nil {
		collector.Add("allowedDomains", err.Error(), settings.AllowedDomains)
	}
	if settings.Enabled {
		validateProviderCommon(
			&collector,
			settings.ClientID,
			settings.ClientSecretSet,
			settings.DisplayName,
			settings.CallbackURL,
		)
	}
	return collector.Err(ErrInvalidInput)
}

func validateGitHub(settings GitHubSettings) error {
	if !settings.Enabled {
		return nil
	}
	var collector appvalidation.Collector
	validateProviderCommon(
		&collector,
		settings.ClientID,
		settings.ClientSecretSet,
		settings.DisplayName,
		settings.CallbackURL,
	)
	return collector.Err(ErrInvalidInput)
}

func validateProviderCommon(
	collector *appvalidation.Collector,
	clientID string,
	clientSecretSet bool,
	displayName string,
	callbackURL *string,
) {
	if clientID == "" {
		collector.Add("clientId", "clientId is required when enabled", clientID)
	}
	if !clientSecretSet {
		collector.Add("clientSecret", "clientSecret is required when enabled", nil)
	}
	if displayName == "" {
		collector.Add("displayName", "displayName is required when enabled", displayName)
	}
	if callbackURL == nil || strings.TrimSpace(*callbackURL) == "" {
		collector.Add("callbackUrl", "BACKEND_BASE_URL is required when enabled", nil)
	}
}

func validateIssuerURL(value string) error {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Hostname() == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("issuerUrl must be an absolute HTTPS URL")
	}
	switch parsed.Scheme {
	case "https":
		return nil
	case "http":
		if isLoopbackHostname(parsed.Hostname()) {
			return nil
		}
	}
	return errors.New("issuerUrl must use HTTPS except for a loopback development host")
}

func isLoopbackHostname(host string) bool {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func normalizeGoogleDomains(values []string) ([]string, error) {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		value := strings.ToLower(strings.TrimSpace(raw))
		if value == "" {
			return nil, errors.New("allowedDomains must not contain empty domains")
		}
		if err := validateDNSName(value); err != nil {
			return nil, fmt.Errorf("allowedDomains contains invalid domain %q", raw)
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}

func validateDNSName(value string) error {
	if len(value) > 253 || strings.HasPrefix(value, ".") || strings.HasSuffix(value, ".") {
		return errors.New("invalid DNS name")
	}
	for _, label := range strings.Split(value, ".") {
		if !dnsLabelPattern.MatchString(label) {
			return errors.New("invalid DNS label")
		}
	}
	return nil
}

func normalizeStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	return &normalized
}

func validateSecretPatch(secret OptionalSecret, field string) error {
	if !secret.Present || secret.Value == nil {
		return nil
	}
	if strings.TrimSpace(*secret.Value) == "" {
		return invalidField(field, field+" must not be empty", *secret.Value)
	}
	return nil
}
