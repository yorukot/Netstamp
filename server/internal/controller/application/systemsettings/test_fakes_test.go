package systemsettings

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	domainsystem "github.com/yorukot/netstamp/internal/domain/system"
)

type settingsAuditEvent struct {
	key    string
	action string
	actor  string
}

type memorySettingsRepository struct {
	settings       map[string]domainsystem.Setting
	auditEvents    []settingsAuditEvent
	auditErr       error
	lockErrors     map[string]error
	readKeyQueries [][]string
	lockAttempts   []string
	upsertAttempts []string
	deleteAttempts []string
	auditAttempts  int
}

func newMemorySettingsRepository() *memorySettingsRepository {
	return &memorySettingsRepository{
		settings:   make(map[string]domainsystem.Setting),
		lockErrors: make(map[string]error),
	}
}

func (r *memorySettingsRepository) GetSystemSettingsByKeys(
	_ context.Context,
	keys []string,
) ([]domainsystem.Setting, error) {
	r.readKeyQueries = append(r.readKeyQueries, append([]string(nil), keys...))
	rows := make([]domainsystem.Setting, 0, len(keys))
	for _, key := range keys {
		if setting, ok := r.settings[key]; ok {
			rows = append(rows, cloneTestSetting(setting))
		}
	}
	return rows, nil
}

func (r *memorySettingsRepository) UpsertSystemSetting(
	_ context.Context,
	setting domainsystem.Setting,
) (domainsystem.Setting, error) {
	r.upsertAttempts = append(r.upsertAttempts, setting.Key)
	r.settings[setting.Key] = cloneTestSetting(setting)
	return cloneTestSetting(setting), nil
}

func (r *memorySettingsRepository) DeleteSystemSetting(_ context.Context, key string) error {
	r.deleteAttempts = append(r.deleteAttempts, key)
	delete(r.settings, key)
	return nil
}

func (r *memorySettingsRepository) CreateSystemSettingAuditEvent(
	_ context.Context,
	key string,
	action string,
	updatedByUserID *string,
) error {
	r.auditAttempts++
	if r.auditErr != nil {
		return r.auditErr
	}
	actor := ""
	if updatedByUserID != nil {
		actor = *updatedByUserID
	}
	r.auditEvents = append(r.auditEvents, settingsAuditEvent{key: key, action: action, actor: actor})
	return nil
}

func (r *memorySettingsRepository) LockSystemSettingsResource(_ context.Context, resource string) error {
	r.lockAttempts = append(r.lockAttempts, resource)
	return r.lockErrors[resource]
}

type memorySettingsSnapshot struct {
	settings    map[string]domainsystem.Setting
	auditEvents []settingsAuditEvent
}

func (r *memorySettingsRepository) snapshot() memorySettingsSnapshot {
	settings := make(map[string]domainsystem.Setting, len(r.settings))
	for key, setting := range r.settings {
		settings[key] = cloneTestSetting(setting)
	}
	return memorySettingsSnapshot{
		settings:    settings,
		auditEvents: append([]settingsAuditEvent(nil), r.auditEvents...),
	}
}

func (r *memorySettingsRepository) restore(snapshot memorySettingsSnapshot) {
	r.settings = snapshot.settings
	r.auditEvents = snapshot.auditEvents
}

type rollbackSettingsTransactor struct {
	repo      *memorySettingsRepository
	calls     int
	commits   int
	rollbacks int
	active    bool
}

func (t *rollbackSettingsTransactor) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	t.calls++
	snapshot := t.repo.snapshot()
	t.active = true
	defer func() { t.active = false }()
	if err := fn(ctx); err != nil {
		t.repo.restore(snapshot)
		t.rollbacks++
		return err
	}
	t.commits++
	return nil
}

type passthroughSettingsTransactor struct{}

func (passthroughSettingsTransactor) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type fakeSystemAdminChecker struct {
	err error
}

func (f fakeSystemAdminChecker) IsSystemAdmin(_ context.Context, userID string) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	return userID == testSystemSettingsAdminID, nil
}

type fakeSecretCipher struct {
	encryptErr    error
	decryptErrors map[string]error
	decryptInputs []string
}

func (f *fakeSecretCipher) Encrypt(plaintext string) ([]byte, []byte, error) {
	if f.encryptErr != nil {
		return nil, nil, f.encryptErr
	}
	return []byte("enc:" + plaintext), []byte("nonce"), nil
}

func (f *fakeSecretCipher) Decrypt(ciphertext, _ []byte) (string, error) {
	value := string(ciphertext)
	f.decryptInputs = append(f.decryptInputs, value)
	if err := f.decryptErrors[value]; err != nil {
		return "", err
	}
	if !strings.HasPrefix(value, "enc:") {
		return "", errors.New("invalid test ciphertext")
	}
	return strings.TrimPrefix(value, "enc:"), nil
}

type fakeOIDCReadinessChecker struct {
	err     error
	issuers []string
	onCheck func(context.Context, string)
}

func (f *fakeOIDCReadinessChecker) Check(ctx context.Context, issuerURL string) error {
	f.issuers = append(f.issuers, issuerURL)
	if f.onCheck != nil {
		f.onCheck(ctx, issuerURL)
	}
	return f.err
}

type fakeOIDCMetadataError struct {
	cause error
}

func (e *fakeOIDCMetadataError) Error() string {
	return e.cause.Error()
}

func (e *fakeOIDCMetadataError) Unwrap() error {
	return e.cause
}

func (*fakeOIDCMetadataError) InvalidOIDCMetadata() {}

func testPublicSetting(t *testing.T, key string, value any) domainsystem.Setting {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encode test setting %q: %v", key, err)
	}
	return domainsystem.Setting{Key: key, Value: encoded}
}

func testSecretSetting(key, plaintext string) domainsystem.Setting {
	return domainsystem.Setting{
		Key:                 key,
		Secret:              true,
		EncryptedValue:      []byte("enc:" + plaintext),
		EncryptedValueNonce: []byte("nonce"),
	}
}

func decodeTestSetting(setting domainsystem.Setting, target any) error {
	return json.Unmarshal(setting.Value, target)
}

func cloneTestSetting(setting domainsystem.Setting) domainsystem.Setting {
	setting.Value = append([]byte(nil), setting.Value...)
	setting.EncryptedValue = append([]byte(nil), setting.EncryptedValue...)
	setting.EncryptedValueNonce = append([]byte(nil), setting.EncryptedValueNonce...)
	if setting.UpdatedByUserID != nil {
		actor := *setting.UpdatedByUserID
		setting.UpdatedByUserID = &actor
	}
	return setting
}
