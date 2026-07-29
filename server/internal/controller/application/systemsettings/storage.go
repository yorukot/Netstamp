package systemsettings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	domainsystem "github.com/yorukot/netstamp/internal/domain/system"
)

type storedOIDCSettings struct {
	Enabled     bool   `json:"enabled"`
	IssuerURL   string `json:"issuerUrl"`
	ClientID    string `json:"clientId"`
	DisplayName string `json:"displayName"`
	JITEnabled  bool   `json:"jitEnabled"`
}

type storedGoogleSettings struct {
	Enabled        bool     `json:"enabled"`
	ClientID       string   `json:"clientId"`
	DisplayName    string   `json:"displayName"`
	JITEnabled     bool     `json:"jitEnabled"`
	AllowedDomains []string `json:"allowedDomains,omitempty"`
}

func (s *storedGoogleSettings) UnmarshalJSON(data []byte) error {
	var wire struct {
		Enabled        bool            `json:"enabled"`
		ClientID       string          `json:"clientId"`
		DisplayName    string          `json:"displayName"`
		JITEnabled     bool            `json:"jitEnabled"`
		AllowedDomains json.RawMessage `json:"allowedDomains"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	var domains []string
	rawDomains := strings.TrimSpace(string(wire.AllowedDomains))
	switch {
	case rawDomains == "":
	case rawDomains == "null":
		return errors.New("allowedDomains must not be null")
	case strings.HasPrefix(rawDomains, `"`):
		var commaSeparated string
		if err := json.Unmarshal(wire.AllowedDomains, &commaSeparated); err != nil {
			return err
		}
		if strings.TrimSpace(commaSeparated) != "" {
			domains = strings.Split(commaSeparated, ",")
		}
	default:
		if err := json.Unmarshal(wire.AllowedDomains, &domains); err != nil {
			return err
		}
	}
	normalized, err := normalizeGoogleDomains(domains)
	if err != nil {
		return err
	}
	*s = storedGoogleSettings{
		Enabled:        wire.Enabled,
		ClientID:       wire.ClientID,
		DisplayName:    wire.DisplayName,
		JITEnabled:     wire.JITEnabled,
		AllowedDomains: normalized,
	}
	return nil
}

type storedGitHubSettings struct {
	Enabled     bool   `json:"enabled"`
	ClientID    string `json:"clientId"`
	DisplayName string `json:"displayName"`
	JITEnabled  bool   `json:"jitEnabled"`
	AllowSignup bool   `json:"allowSignup"`
}

func (s *Service) settingsByKeys(ctx context.Context, keys []string) (map[string]domainsystem.Setting, error) {
	if s.repo == nil {
		return nil, errors.New("system settings repository is unavailable")
	}
	rows, err := s.repo.GetSystemSettingsByKeys(ctx, keys)
	if err != nil {
		return nil, err
	}
	result := make(map[string]domainsystem.Setting, len(rows))
	for _, row := range rows {
		result[row.Key] = row
	}
	return result, nil
}

func applyPublicSetting[T any](settings map[string]domainsystem.Setting, key string, target *T) error {
	row, ok := settings[key]
	if !ok {
		return nil
	}
	value := bytes.TrimSpace(row.Value)
	if row.Secret || len(value) == 0 || bytes.Equal(value, []byte("null")) {
		return fmt.Errorf("system setting %q has an invalid public value", key)
	}
	if err := json.Unmarshal(value, target); err != nil {
		return fmt.Errorf("decode system setting %q: %w", key, err)
	}
	return nil
}

func secretIsSet(settings map[string]domainsystem.Setting, key string) bool {
	_, ok := settings[key]
	return ok
}

func (s *Service) decryptSecret(settings map[string]domainsystem.Setting, key string) (string, error) {
	row, ok := settings[key]
	if !ok {
		return "", nil
	}
	if !row.Secret || len(row.EncryptedValue) == 0 || len(row.EncryptedValueNonce) == 0 {
		return "", fmt.Errorf("system setting %q has an invalid secret value", key)
	}
	if s.cipher == nil {
		return "", errors.New("system settings secret cipher is unavailable")
	}
	value, err := s.cipher.Decrypt(row.EncryptedValue, row.EncryptedValueNonce)
	if err != nil {
		return "", fmt.Errorf("decrypt system setting %q: %w", key, err)
	}
	if value == "" {
		return "", fmt.Errorf("decrypt system setting %q: secret value is empty", key)
	}
	return value, nil
}

func publicStoredSetting(key string, value any, actor *string) (domainsystem.Setting, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return domainsystem.Setting{}, fmt.Errorf("encode system setting %q: %w", key, err)
	}
	return domainsystem.Setting{
		Key:             key,
		Value:           encoded,
		UpdatedByUserID: actor,
	}, nil
}

func (s *Service) secretStoredSetting(key, value string, actor *string) (domainsystem.Setting, error) {
	if s.cipher == nil {
		return domainsystem.Setting{}, errors.New("system settings secret cipher is unavailable")
	}
	ciphertext, nonce, err := s.cipher.Encrypt(value)
	if err != nil {
		return domainsystem.Setting{}, fmt.Errorf("encrypt system setting %q: %w", key, err)
	}
	return domainsystem.Setting{
		Key:                 key,
		EncryptedValue:      ciphertext,
		EncryptedValueNonce: nonce,
		Secret:              true,
		UpdatedByUserID:     actor,
	}, nil
}

func (s *Service) upsertPublic(ctx context.Context, key string, value any, actor *string) error {
	setting, err := publicStoredSetting(key, value, actor)
	if err != nil {
		return err
	}
	if _, err := s.repo.UpsertSystemSetting(ctx, setting); err != nil {
		return err
	}
	return s.repo.CreateSystemSettingAuditEvent(ctx, key, auditActionUpdate, actor)
}

func (s *Service) upsertSecret(ctx context.Context, key, value string, actor *string) error {
	setting, err := s.secretStoredSetting(key, value, actor)
	if err != nil {
		return err
	}
	if _, err := s.repo.UpsertSystemSetting(ctx, setting); err != nil {
		return err
	}
	return s.repo.CreateSystemSettingAuditEvent(ctx, key, auditActionUpdate, actor)
}

func (s *Service) deleteSecret(ctx context.Context, key string, actor *string) error {
	if err := s.repo.DeleteSystemSetting(ctx, key); err != nil {
		return err
	}
	return s.repo.CreateSystemSettingAuditEvent(ctx, key, auditActionClear, actor)
}
