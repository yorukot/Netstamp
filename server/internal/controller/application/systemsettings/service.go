package systemsettings

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const defaultReadinessTimeout = 5 * time.Second

type Service struct {
	repo             Repository
	admins           SystemAdminChecker
	cipher           SecretCipher
	readiness        OIDCReadinessChecker
	transactor       Transactor
	defaults         Defaults
	callbackBaseURL  string
	readinessTimeout time.Duration
	smtpUsers        SMTPTestUserRepository
	smtpTester       SMTPTester
}

type unavailableTransactor struct{}

func (unavailableTransactor) WithinTx(context.Context, func(context.Context) error) error {
	return errors.New("system settings transactor is unavailable")
}

// callbackBaseURL is the versioned external-auth base, for example
// https://netstamp.example/api/v1/auth/external.
func NewService(
	repo Repository,
	admins SystemAdminChecker,
	cipher SecretCipher,
	readiness OIDCReadinessChecker,
	defaults Defaults,
	callbackBaseURL string,
	transactors ...Transactor,
) *Service {
	normalizeDefaults(&defaults)
	var transactor Transactor = unavailableTransactor{}
	if len(transactors) > 0 && transactors[0] != nil {
		transactor = transactors[0]
	}
	return &Service{
		repo:             repo,
		admins:           admins,
		cipher:           cipher,
		readiness:        readiness,
		transactor:       transactor,
		defaults:         defaults,
		callbackBaseURL:  strings.TrimRight(strings.TrimSpace(callbackBaseURL), "/"),
		readinessTimeout: defaultReadinessTimeout,
	}
}

func normalizeDefaults(defaults *Defaults) {
	if defaults.Access == (AccessSettings{}) {
		defaults.Access = AccessSettings{
			AccountCreationEnabled:   true,
			ProjectCreationEnabled:   true,
			CredentialChangesEnabled: true,
		}
	}
	if defaults.SMTP.Port == 0 {
		defaults.SMTP.Port = 587
	}
	if defaults.SMTP.TLSMode == "" {
		defaults.SMTP.TLSMode = "starttls"
	}
	if defaults.SMTP.TimeoutSeconds == 0 {
		defaults.SMTP.TimeoutSeconds = 10
	}
	if defaults.OIDC.DisplayName == "" {
		defaults.OIDC.DisplayName = "Single sign-on"
	}
	if defaults.Google.DisplayName == "" {
		defaults.Google.DisplayName = "Google"
	}
	if defaults.GitHub.DisplayName == "" {
		defaults.GitHub.DisplayName = "GitHub"
	}
}

func (s *Service) ConfigureSMTPTest(users SMTPTestUserRepository, tester SMTPTester) {
	s.smtpUsers = users
	s.smtpTester = tester
}

func (s *Service) requireSystemAdmin(ctx context.Context, userID string) error {
	if strings.TrimSpace(userID) == "" || s.admins == nil {
		return ErrForbidden
	}
	ok, err := s.admins.IsSystemAdmin(ctx, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

func (s *Service) GetAccess(ctx context.Context, input GetAccessInput) (Versioned[AccessSettings], error) {
	return getVersionedSetting(ctx, s, input.CurrentUserID, ResourceAccess, s.EffectiveAccess)
}

func (s *Service) GetSMTP(ctx context.Context, input GetSMTPInput) (Versioned[SMTPSettings], error) {
	return getVersionedSetting(ctx, s, input.CurrentUserID, ResourceSMTP, s.effectiveSMTPView)
}

func (s *Service) GetOIDC(ctx context.Context, input GetOIDCInput) (Versioned[OIDCSettings], error) {
	return getVersionedSetting(ctx, s, input.CurrentUserID, ResourceOIDC, s.effectiveOIDCView)
}

func (s *Service) GetGoogle(ctx context.Context, input GetGoogleInput) (Versioned[GoogleSettings], error) {
	return getVersionedSetting(ctx, s, input.CurrentUserID, ResourceGoogle, s.effectiveGoogleView)
}

func (s *Service) GetGitHub(ctx context.Context, input GetGitHubInput) (Versioned[GitHubSettings], error) {
	return getVersionedSetting(ctx, s, input.CurrentUserID, ResourceGitHub, s.effectiveGitHubView)
}

func getVersionedSetting[T any](
	ctx context.Context,
	service *Service,
	currentUserID string,
	resource Resource,
	load func(context.Context) (T, error),
) (Versioned[T], error) {
	if service == nil || service.repo == nil {
		return Versioned[T]{}, errors.New("system settings repository is unavailable")
	}
	if adminErr := service.requireSystemAdmin(ctx, currentUserID); adminErr != nil {
		return Versioned[T]{}, adminErr
	}
	var result Versioned[T]
	txErr := service.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		revision, revisionErr := service.repo.LockSystemSettingRevision(txCtx, string(resource))
		if revisionErr != nil {
			return revisionErr
		}
		value, loadErr := load(txCtx)
		if loadErr != nil {
			return loadErr
		}
		result = Versioned[T]{Value: value, Revision: revision}
		return nil
	})
	return result, txErr
}

func (s *Service) revision(ctx context.Context, resource Resource) (int64, error) {
	if s.repo == nil {
		return 0, errors.New("system settings repository is unavailable")
	}
	return s.repo.GetSystemSettingRevision(ctx, string(resource))
}

func (s *Service) EffectiveAccess(ctx context.Context) (AccessSettings, error) {
	value := s.defaults.Access
	rows, err := s.settingsByKeys(ctx, accessKeys)
	if err != nil {
		return AccessSettings{}, err
	}
	for key, target := range map[string]*bool{
		keyAccountCreationEnabled:    &value.AccountCreationEnabled,
		keyEmailVerificationRequired: &value.EmailVerificationRequired,
		keyProjectCreationEnabled:    &value.ProjectCreationEnabled,
		keyCredentialChangesEnabled:  &value.CredentialChangesEnabled,
	} {
		if err := applyPublicSetting(rows, key, target); err != nil {
			return AccessSettings{}, err
		}
	}
	return value, nil
}

func (s *Service) effectiveSMTPView(ctx context.Context) (SMTPSettings, error) {
	rows, err := s.settingsByKeys(ctx, smtpKeys)
	if err != nil {
		return SMTPSettings{}, err
	}
	return s.smtpViewFromRows(rows)
}

func (s *Service) smtpViewFromRows(rows map[string]StoredSetting) (SMTPSettings, error) {
	value := s.defaults.SMTP
	for key, target := range map[string]*string{
		keySMTPHost:     &value.Host,
		keySMTPUsername: &value.Username,
		keySMTPFrom:     &value.From,
		keySMTPTLSMode:  &value.TLSMode,
	} {
		if err := applyPublicSetting(rows, key, target); err != nil {
			return SMTPSettings{}, err
		}
	}
	for key, target := range map[string]*int32{
		keySMTPPort:           &value.Port,
		keySMTPTimeoutSeconds: &value.TimeoutSeconds,
	} {
		if err := applyPublicSetting(rows, key, target); err != nil {
			return SMTPSettings{}, err
		}
	}
	value.Host = strings.TrimSpace(value.Host)
	value.Username = strings.TrimSpace(value.Username)
	value.From = strings.TrimSpace(value.From)
	value.TLSMode = strings.TrimSpace(value.TLSMode)
	value.PasswordSet = secretIsSet(rows, keySMTPPassword)
	value.Configured = smtpDeliveryConfigured(value)
	return value, nil
}

func (s *Service) EffectiveSMTP(ctx context.Context) (SMTPRuntimeSettings, error) {
	rows, err := s.settingsByKeys(ctx, smtpKeys)
	if err != nil {
		return SMTPRuntimeSettings{}, err
	}
	view, err := s.smtpViewFromRows(rows)
	if err != nil {
		return SMTPRuntimeSettings{}, err
	}
	password, err := s.decryptSecret(rows, keySMTPPassword)
	if err != nil {
		return SMTPRuntimeSettings{}, err
	}
	if validationErr := validateSMTP(view); validationErr != nil {
		return SMTPRuntimeSettings{}, validationErr
	}
	return SMTPRuntimeSettings{
		Host: view.Host, Port: view.Port, Username: view.Username, Password: password,
		From: view.From, TLSMode: view.TLSMode, TimeoutSeconds: view.TimeoutSeconds,
	}, nil
}

func (s *Service) effectiveOIDCView(ctx context.Context) (OIDCSettings, error) {
	rows, err := s.settingsByKeys(ctx, oidcKeys)
	if err != nil {
		return OIDCSettings{}, err
	}
	return s.oidcViewFromRows(rows)
}

func (s *Service) oidcViewFromRows(rows map[string]StoredSetting) (OIDCSettings, error) {
	value := s.defaults.OIDC
	stored := storedOIDCSettings{
		Enabled: value.Enabled, IssuerURL: value.IssuerURL, ClientID: value.ClientID,
		DisplayName: value.DisplayName, JITEnabled: value.JITEnabled,
	}
	if err := applyPublicSetting(rows, keyOIDCSettings, &stored); err != nil {
		return OIDCSettings{}, err
	}
	return OIDCSettings{
		Enabled: stored.Enabled, IssuerURL: strings.TrimSpace(stored.IssuerURL),
		ClientID:        strings.TrimSpace(stored.ClientID),
		ClientSecretSet: secretIsSet(rows, keyOIDCClientSecret),
		DisplayName:     strings.TrimSpace(stored.DisplayName),
		JITEnabled:      stored.JITEnabled, CallbackURL: s.callbackURL("oidc"),
	}, nil
}

func (s *Service) effectiveGoogleView(ctx context.Context) (GoogleSettings, error) {
	rows, err := s.settingsByKeys(ctx, googleKeys)
	if err != nil {
		return GoogleSettings{}, err
	}
	return s.googleViewFromRows(rows)
}

func (s *Service) googleViewFromRows(rows map[string]StoredSetting) (GoogleSettings, error) {
	value := s.defaults.Google
	stored := storedGoogleSettings{
		Enabled: value.Enabled, ClientID: value.ClientID, DisplayName: value.DisplayName,
		JITEnabled: value.JITEnabled, AllowedDomains: cloneStrings(value.AllowedDomains),
	}
	if err := applyPublicSetting(rows, keyGoogleSettings, &stored); err != nil {
		return GoogleSettings{}, err
	}
	return GoogleSettings{
		Enabled: stored.Enabled, ClientID: strings.TrimSpace(stored.ClientID),
		ClientSecretSet: secretIsSet(rows, keyGoogleClientSecret),
		DisplayName:     strings.TrimSpace(stored.DisplayName),
		JITEnabled:      stored.JITEnabled, AllowedDomains: cloneStrings(stored.AllowedDomains),
		CallbackURL: s.callbackURL("google"),
	}, nil
}

func (s *Service) effectiveGitHubView(ctx context.Context) (GitHubSettings, error) {
	rows, err := s.settingsByKeys(ctx, gitHubKeys)
	if err != nil {
		return GitHubSettings{}, err
	}
	return s.gitHubViewFromRows(rows)
}

func (s *Service) gitHubViewFromRows(rows map[string]StoredSetting) (GitHubSettings, error) {
	value := s.defaults.GitHub
	stored := storedGitHubSettings{
		Enabled: value.Enabled, ClientID: value.ClientID, DisplayName: value.DisplayName,
		JITEnabled: value.JITEnabled, AllowSignup: value.AllowSignup,
	}
	if err := applyPublicSetting(rows, keyGitHubSettings, &stored); err != nil {
		return GitHubSettings{}, err
	}
	return GitHubSettings{
		Enabled: stored.Enabled, ClientID: strings.TrimSpace(stored.ClientID),
		ClientSecretSet: secretIsSet(rows, keyGitHubClientSecret),
		DisplayName:     strings.TrimSpace(stored.DisplayName),
		JITEnabled:      stored.JITEnabled, AllowSignup: stored.AllowSignup, CallbackURL: s.callbackURL("github"),
	}, nil
}

func (s *Service) EffectiveOIDC(ctx context.Context) (OIDCRuntimeSettings, error) {
	rows, err := s.settingsByKeys(ctx, oidcKeys)
	if err != nil {
		return OIDCRuntimeSettings{}, err
	}
	view, err := s.oidcViewFromRows(rows)
	if err != nil {
		return OIDCRuntimeSettings{}, err
	}
	if !view.Enabled {
		return OIDCRuntimeSettings{
			DisplayName: view.DisplayName,
			JITEnabled:  view.JITEnabled,
		}, nil
	}
	if validationErr := validateOIDC(view); validationErr != nil {
		return OIDCRuntimeSettings{}, validationErr
	}
	secret, err := s.decryptSecret(rows, keyOIDCClientSecret)
	if err != nil {
		return OIDCRuntimeSettings{}, err
	}
	return OIDCRuntimeSettings{
		Enabled: view.Enabled, IssuerURL: view.IssuerURL, ClientID: view.ClientID,
		ClientSecret: secret, DisplayName: view.DisplayName, JITEnabled: view.JITEnabled,
	}, nil
}

func (s *Service) EffectiveGoogle(ctx context.Context) (GoogleRuntimeSettings, error) {
	rows, err := s.settingsByKeys(ctx, googleKeys)
	if err != nil {
		return GoogleRuntimeSettings{}, err
	}
	view, err := s.googleViewFromRows(rows)
	if err != nil {
		return GoogleRuntimeSettings{}, err
	}
	if !view.Enabled {
		return GoogleRuntimeSettings{
			DisplayName:    view.DisplayName,
			JITEnabled:     view.JITEnabled,
			AllowedDomains: cloneStrings(view.AllowedDomains),
		}, nil
	}
	if validationErr := validateGoogle(view); validationErr != nil {
		return GoogleRuntimeSettings{}, validationErr
	}
	secret, err := s.decryptSecret(rows, keyGoogleClientSecret)
	if err != nil {
		return GoogleRuntimeSettings{}, err
	}
	return GoogleRuntimeSettings{
		Enabled: view.Enabled, ClientID: view.ClientID, ClientSecret: secret,
		DisplayName: view.DisplayName, JITEnabled: view.JITEnabled,
		AllowedDomains: cloneStrings(view.AllowedDomains),
	}, nil
}

func (s *Service) EffectiveGitHub(ctx context.Context) (GitHubRuntimeSettings, error) {
	rows, err := s.settingsByKeys(ctx, gitHubKeys)
	if err != nil {
		return GitHubRuntimeSettings{}, err
	}
	view, err := s.gitHubViewFromRows(rows)
	if err != nil {
		return GitHubRuntimeSettings{}, err
	}
	if !view.Enabled {
		return GitHubRuntimeSettings{
			DisplayName: view.DisplayName,
			JITEnabled:  view.JITEnabled,
			AllowSignup: view.AllowSignup,
		}, nil
	}
	if validationErr := validateGitHub(view); validationErr != nil {
		return GitHubRuntimeSettings{}, validationErr
	}
	secret, err := s.decryptSecret(rows, keyGitHubClientSecret)
	if err != nil {
		return GitHubRuntimeSettings{}, err
	}
	return GitHubRuntimeSettings{
		Enabled: view.Enabled, ClientID: view.ClientID, ClientSecret: secret,
		DisplayName: view.DisplayName, JITEnabled: view.JITEnabled, AllowSignup: view.AllowSignup,
	}, nil
}

func (s *Service) callbackURL(provider string) *string {
	if s.callbackBaseURL == "" {
		return nil
	}
	value := s.callbackBaseURL + "/" + provider + "/callback"
	parsed, err := url.Parse(value)
	if err != nil ||
		!parsed.IsAbs() ||
		parsed.Hostname() == "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.User != nil ||
		parsed.RawQuery != "" ||
		parsed.Fragment != "" {
		return nil
	}
	return &value
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}

func (s *Service) AccountCreationEnabled(ctx context.Context) (bool, error) {
	value, err := s.EffectiveAccess(ctx)
	return value.AccountCreationEnabled, err
}

func (s *Service) EmailVerificationRequired(ctx context.Context) (bool, error) {
	value, err := s.EffectiveAccess(ctx)
	return value.EmailVerificationRequired, err
}

func (s *Service) ProjectCreationEnabled(ctx context.Context) (bool, error) {
	value, err := s.EffectiveAccess(ctx)
	return value.ProjectCreationEnabled, err
}

func (s *Service) CredentialChangesEnabled(ctx context.Context) (bool, error) {
	value, err := s.EffectiveAccess(ctx)
	return value.CredentialChangesEnabled, err
}

func (s *Service) TestSMTP(ctx context.Context, input TestSMTPInput) error {
	ctx, flow := s.startSettingsFlow(ctx, "test", ResourceSMTP, input.CurrentUserID, input.ExpectedRevision)
	defer flow.end()
	if err := flow.authorize(ctx); err != nil {
		return err
	}
	var snapshot SMTPRuntimeSettings
	if snapshotErr := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		revision, lockErr := s.repo.LockSystemSettingRevision(txCtx, string(ResourceSMTP))
		if lockErr != nil {
			return lockErr
		}
		if revisionErr := flow.requireRevision(revision); revisionErr != nil {
			return revisionErr
		}
		rows, loadErr := s.settingsByKeys(txCtx, smtpKeys)
		if loadErr != nil {
			return loadErr
		}
		settings, viewErr := s.smtpViewFromRows(rows)
		if viewErr != nil {
			return viewErr
		}
		if validationErr := validateSMTP(settings); validationErr != nil {
			return validationErr
		}
		if !settings.Configured {
			return invalidField("smtp", "SMTP must be configured before sending a test", nil)
		}
		password, decryptErr := s.decryptSecret(rows, keySMTPPassword)
		if decryptErr != nil {
			return decryptErr
		}
		snapshot = SMTPRuntimeSettings{
			Host: settings.Host, Port: settings.Port, Username: settings.Username, Password: password,
			From: settings.From, TLSMode: settings.TLSMode, TimeoutSeconds: settings.TimeoutSeconds,
		}
		return nil
	}); snapshotErr != nil {
		return snapshotErr
	}
	if s.smtpUsers == nil || s.smtpTester == nil {
		return ErrSMTPTestFailed
	}
	user, err := s.smtpUsers.GetUserByID(ctx, input.CurrentUserID)
	if err != nil {
		return err
	}
	if user.DisabledAt != nil {
		return ErrForbidden
	}
	if sendErr := s.smtpTester.SendTestEmail(ctx, user.Email, snapshot); sendErr != nil {
		return fmt.Errorf("%w: %w", ErrSMTPTestFailed, sendErr)
	}
	return nil
}
