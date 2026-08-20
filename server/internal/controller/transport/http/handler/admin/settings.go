package admin

import (
	"bytes"
	"encoding/json"
	"errors"
)

type optionalNullableString struct {
	Present bool
	Value   *string
}

type optionalNonNull[T any] struct {
	Value *T
}

func (value *optionalNonNull[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("settings patch field must not be null")
	}

	var decoded T
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	value.Value = &decoded
	return nil
}

func (value *optionalNullableString) UnmarshalJSON(data []byte) error {
	value.Present = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		value.Value = nil
		return nil
	}

	var decoded string
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	value.Value = &decoded
	return nil
}

type accessSettingsPatchBody struct {
	AccountCreationEnabled    optionalNonNull[bool] `json:"accountCreationEnabled"`
	EmailVerificationRequired optionalNonNull[bool] `json:"emailVerificationRequired"`
	ProjectCreationEnabled    optionalNonNull[bool] `json:"projectCreationEnabled"`
	CredentialChangesEnabled  optionalNonNull[bool] `json:"credentialChangesEnabled"`
}

type updatesSettingsPatchBody struct {
	CheckForUpdates optionalNonNull[bool] `json:"checkForUpdates"`
}

type smtpSettingsPatchBody struct {
	Host           optionalNonNull[string] `json:"host"`
	Port           optionalNonNull[int32]  `json:"port"`
	Username       optionalNonNull[string] `json:"username"`
	Password       optionalNullableString  `json:"password"`
	From           optionalNonNull[string] `json:"from"`
	TLSMode        optionalNonNull[string] `json:"tlsMode"`
	TimeoutSeconds optionalNonNull[int32]  `json:"timeoutSeconds"`
}

type providerSettingsPatchBody struct {
	Enabled      optionalNonNull[bool]   `json:"enabled"`
	ClientID     optionalNonNull[string] `json:"clientId"`
	ClientSecret optionalNullableString  `json:"clientSecret"`
	DisplayName  optionalNonNull[string] `json:"displayName"`
	JITEnabled   optionalNonNull[bool]   `json:"jitEnabled"`
}

type oidcProviderSettingsPatchBody struct {
	providerSettingsPatchBody
	IssuerURL optionalNonNull[string] `json:"issuerUrl"`
}

type googleProviderSettingsPatchBody struct {
	providerSettingsPatchBody
	AllowedDomains optionalNonNull[[]string] `json:"allowedDomains"`
}

type githubProviderSettingsPatchBody struct {
	providerSettingsPatchBody
	AllowSignup optionalNonNull[bool] `json:"allowSignup"`
}

type settingsEnvelope[T any] struct {
	Settings T `json:"settings"`
}

type accessSettingsResponseBody struct {
	AccountCreationEnabled    bool `json:"accountCreationEnabled"`
	EmailVerificationRequired bool `json:"emailVerificationRequired"`
	ProjectCreationEnabled    bool `json:"projectCreationEnabled"`
	CredentialChangesEnabled  bool `json:"credentialChangesEnabled"`
}

type updatesSettingsResponseBody struct {
	CheckForUpdates bool `json:"checkForUpdates"`
}

type smtpSettingsResponseBody struct {
	Host           string `json:"host"`
	Port           int32  `json:"port"`
	Username       string `json:"username"`
	PasswordSet    bool   `json:"passwordSet"`
	From           string `json:"from"`
	TLSMode        string `json:"tlsMode"`
	TimeoutSeconds int32  `json:"timeoutSeconds"`
	Configured     bool   `json:"configured"`
}

type providerSettingsResponseBody struct {
	Enabled         bool    `json:"enabled"`
	ClientID        string  `json:"clientId"`
	ClientSecretSet bool    `json:"clientSecretSet"`
	DisplayName     string  `json:"displayName"`
	JITEnabled      bool    `json:"jitEnabled"`
	CallbackURL     *string `json:"callbackUrl,omitempty"`
}

type oidcProviderSettingsResponseBody struct {
	providerSettingsResponseBody
	IssuerURL string `json:"issuerUrl"`
}

type googleProviderSettingsResponseBody struct {
	providerSettingsResponseBody
	AllowedDomains []string `json:"allowedDomains"`
}

type githubProviderSettingsResponseBody struct {
	providerSettingsResponseBody
	AllowSignup bool `json:"allowSignup"`
}
