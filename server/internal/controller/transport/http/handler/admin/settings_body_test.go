package admin

import (
	"encoding/json"
	"testing"
)

func TestSMTPPasswordPatchDistinguishesOmittedNullAndValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		body        string
		wantPresent bool
		wantValue   *string
	}{
		{name: "omitted", body: `{}`},
		{name: "clear", body: `{"password":null}`, wantPresent: true},
		{name: "replace", body: `{"password":"secret"}`, wantPresent: true, wantValue: stringPointer("secret")},
		{name: "empty remains present for application validation", body: `{"password":""}`, wantPresent: true, wantValue: stringPointer("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var body smtpSettingsPatchBody
			if err := json.Unmarshal([]byte(tt.body), &body); err != nil {
				t.Fatalf("decode SMTP patch: %v", err)
			}
			if body.Password.Present != tt.wantPresent {
				t.Fatalf("expected present=%t, got %t", tt.wantPresent, body.Password.Present)
			}
			if !equalOptionalString(body.Password.Value, tt.wantValue) {
				t.Fatalf("expected value %#v, got %#v", tt.wantValue, body.Password.Value)
			}
		})
	}
}

func TestProviderClientSecretPatchDistinguishesOmittedNullAndValue(t *testing.T) {
	t.Parallel()

	var body googleProviderSettingsPatchBody
	if err := json.Unmarshal([]byte(`{"clientSecret":null,"allowedDomains":["example.com"]}`), &body); err != nil {
		t.Fatalf("decode provider patch: %v", err)
	}
	if !body.ClientSecret.Present || body.ClientSecret.Value != nil {
		t.Fatalf("expected an explicit client secret clear, got %#v", body.ClientSecret)
	}
	if body.AllowedDomains.Value == nil || len(*body.AllowedDomains.Value) != 1 || (*body.AllowedDomains.Value)[0] != "example.com" {
		t.Fatalf("unexpected allowed domains: %#v", body.AllowedDomains.Value)
	}
}

func TestGoogleAllowedDomainsPatchPreservesExplicitEmptyList(t *testing.T) {
	t.Parallel()

	var body googleProviderSettingsPatchBody
	if err := json.Unmarshal([]byte(`{"allowedDomains":[]}`), &body); err != nil {
		t.Fatalf("decode provider patch: %v", err)
	}
	if body.AllowedDomains.Value == nil {
		t.Fatal("expected explicit empty allowedDomains patch to remain present")
	}
	if len(*body.AllowedDomains.Value) != 0 {
		t.Fatalf("expected empty allowedDomains, got %#v", *body.AllowedDomains.Value)
	}
}

func TestSettingsPatchRejectsNullForEveryNonSecretField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		body   string
		target func() any
	}{
		{name: "access account creation", body: `{"accountCreationEnabled":null}`, target: func() any { return &accessSettingsPatchBody{} }},
		{name: "access email verification", body: `{"emailVerificationRequired":null}`, target: func() any { return &accessSettingsPatchBody{} }},
		{name: "access project creation", body: `{"projectCreationEnabled":null}`, target: func() any { return &accessSettingsPatchBody{} }},
		{name: "access credential changes", body: `{"credentialChangesEnabled":null}`, target: func() any { return &accessSettingsPatchBody{} }},
		{name: "SMTP host", body: `{"host":null}`, target: func() any { return &smtpSettingsPatchBody{} }},
		{name: "SMTP port", body: `{"port":null}`, target: func() any { return &smtpSettingsPatchBody{} }},
		{name: "SMTP username", body: `{"username":null}`, target: func() any { return &smtpSettingsPatchBody{} }},
		{name: "SMTP from", body: `{"from":null}`, target: func() any { return &smtpSettingsPatchBody{} }},
		{name: "SMTP TLS mode", body: `{"tlsMode":null}`, target: func() any { return &smtpSettingsPatchBody{} }},
		{name: "SMTP timeout", body: `{"timeoutSeconds":null}`, target: func() any { return &smtpSettingsPatchBody{} }},
		{name: "provider enabled", body: `{"enabled":null}`, target: func() any { return &oidcProviderSettingsPatchBody{} }},
		{name: "provider client ID", body: `{"clientId":null}`, target: func() any { return &oidcProviderSettingsPatchBody{} }},
		{name: "provider display name", body: `{"displayName":null}`, target: func() any { return &oidcProviderSettingsPatchBody{} }},
		{name: "provider JIT", body: `{"jitEnabled":null}`, target: func() any { return &oidcProviderSettingsPatchBody{} }},
		{name: "OIDC issuer", body: `{"issuerUrl":null}`, target: func() any { return &oidcProviderSettingsPatchBody{} }},
		{name: "Google allowed domains", body: `{"allowedDomains":null}`, target: func() any { return &googleProviderSettingsPatchBody{} }},
		{name: "GitHub allow signup", body: `{"allowSignup":null}`, target: func() any { return &githubProviderSettingsPatchBody{} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := json.Unmarshal([]byte(tt.body), tt.target()); err == nil {
				t.Fatalf("expected %s to reject null", tt.body)
			}
		})
	}
}

func stringPointer(value string) *string {
	return &value
}

func equalOptionalString(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
