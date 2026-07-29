package systemsettings

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

func (s *Service) UpdateOIDC(ctx context.Context, input UpdateOIDCInput) (Versioned[OIDCSettings], error) {
	ctx, flow := s.startSettingsFlow(ctx, "update", ResourceOIDC, input.CurrentUserID, input.ExpectedRevision)
	defer flow.end()
	if authorizeErr := flow.authorize(ctx); authorizeErr != nil {
		return Versioned[OIDCSettings]{}, authorizeErr
	}
	if secretErr := validateSecretPatch(input.ClientSecret, "clientSecret"); secretErr != nil {
		return Versioned[OIDCSettings]{}, secretErr
	}
	normalizeOIDCInput(&input)
	if prepareErr := s.prepareOIDCUpdate(ctx, flow, input, false); prepareErr != nil {
		return Versioned[OIDCSettings]{}, prepareErr
	}

	return runProviderUpdate(ctx, s, flow, providerUpdateOperations[OIDCSettings]{
		load:  s.effectiveOIDCView,
		apply: func(current OIDCSettings) OIDCSettings { return applyOIDCPatch(current, input) },
		validate: func(txCtx context.Context, next OIDCSettings) error {
			return s.validateOIDCCandidate(txCtx, next, input.ClientSecret)
		},
		noChange: func(current, next OIDCSettings) bool {
			return oidcPatchIsNoop(current, next, input.ClientSecret)
		},
		persist: func(txCtx context.Context, current, next OIDCSettings) error {
			return s.persistOIDC(txCtx, current, next, input.ClientSecret, &input.CurrentUserID)
		},
	})
}

func (s *Service) UpdateGoogle(ctx context.Context, input UpdateGoogleInput) (Versioned[GoogleSettings], error) {
	ctx, flow := s.startSettingsFlow(ctx, "update", ResourceGoogle, input.CurrentUserID, input.ExpectedRevision)
	defer flow.end()
	if authorizeErr := flow.authorize(ctx); authorizeErr != nil {
		return Versioned[GoogleSettings]{}, authorizeErr
	}
	if secretErr := validateSecretPatch(input.ClientSecret, "clientSecret"); secretErr != nil {
		return Versioned[GoogleSettings]{}, secretErr
	}
	if normalizeErr := normalizeGoogleInput(&input); normalizeErr != nil {
		return Versioned[GoogleSettings]{}, normalizeErr
	}

	return runProviderUpdate(ctx, s, flow, providerUpdateOperations[GoogleSettings]{
		load:  s.effectiveGoogleView,
		apply: func(current GoogleSettings) GoogleSettings { return applyGooglePatch(current, input) },
		validate: func(txCtx context.Context, next GoogleSettings) error {
			return s.validateGoogleCandidate(txCtx, next, input.ClientSecret)
		},
		noChange: func(current, next GoogleSettings) bool {
			return googlePatchIsNoop(current, next, input.ClientSecret)
		},
		persist: func(txCtx context.Context, current, next GoogleSettings) error {
			return s.persistGoogle(txCtx, current, next, input.ClientSecret, &input.CurrentUserID)
		},
	})
}

func (s *Service) UpdateGitHub(ctx context.Context, input UpdateGitHubInput) (Versioned[GitHubSettings], error) {
	ctx, flow := s.startSettingsFlow(ctx, "update", ResourceGitHub, input.CurrentUserID, input.ExpectedRevision)
	defer flow.end()
	if authorizeErr := flow.authorize(ctx); authorizeErr != nil {
		return Versioned[GitHubSettings]{}, authorizeErr
	}
	if secretErr := validateSecretPatch(input.ClientSecret, "clientSecret"); secretErr != nil {
		return Versioned[GitHubSettings]{}, secretErr
	}
	normalizeGitHubInput(&input)

	return runProviderUpdate(ctx, s, flow, providerUpdateOperations[GitHubSettings]{
		load:  s.effectiveGitHubView,
		apply: func(current GitHubSettings) GitHubSettings { return applyGitHubPatch(current, input) },
		validate: func(txCtx context.Context, next GitHubSettings) error {
			return s.validateGitHubCandidate(txCtx, next, input.ClientSecret)
		},
		noChange: func(current, next GitHubSettings) bool {
			return gitHubPatchIsNoop(current, next, input.ClientSecret)
		},
		persist: func(txCtx context.Context, current, next GitHubSettings) error {
			return s.persistGitHub(txCtx, current, next, input.ClientSecret, &input.CurrentUserID)
		},
	})
}

type providerUpdateOperations[T any] struct {
	load     func(context.Context) (T, error)
	apply    func(T) T
	validate func(context.Context, T) error
	noChange func(T, T) bool
	persist  func(context.Context, T, T) error
}

func runProviderUpdate[T any](
	ctx context.Context,
	service *Service,
	flow settingsFlow,
	operations providerUpdateOperations[T],
) (Versioned[T], error) {
	var result Versioned[T]
	txErr := service.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		revision, lockErr := service.repo.LockSystemSettingRevision(txCtx, string(flow.resource))
		if lockErr != nil {
			return lockErr
		}
		if revisionErr := flow.requireRevision(revision); revisionErr != nil {
			return revisionErr
		}
		current, loadErr := operations.load(txCtx)
		if loadErr != nil {
			return loadErr
		}
		next := operations.apply(current)
		if validationErr := operations.validate(txCtx, next); validationErr != nil {
			return validationErr
		}
		if operations.noChange(current, next) {
			result = Versioned[T]{Value: current, Revision: revision}
			return nil
		}
		if persistErr := operations.persist(txCtx, current, next); persistErr != nil {
			return persistErr
		}
		revision, bumpErr := service.repo.BumpSystemSettingRevision(txCtx, string(flow.resource))
		if bumpErr != nil {
			return bumpErr
		}
		result = Versioned[T]{Value: next, Revision: revision}
		return nil
	})
	return result, txErr
}

func (s *Service) prepareOIDCUpdate(
	ctx context.Context,
	flow settingsFlow,
	input UpdateOIDCInput,
	forceReadiness bool,
) error {
	var current OIDCSettings
	var next OIDCSettings
	snapshotErr := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		revision, lockErr := s.repo.LockSystemSettingRevision(txCtx, string(ResourceOIDC))
		if lockErr != nil {
			return lockErr
		}
		if revisionErr := flow.requireRevision(revision); revisionErr != nil {
			return revisionErr
		}
		var loadErr error
		current, loadErr = s.effectiveOIDCView(txCtx)
		if loadErr != nil {
			return loadErr
		}
		next = applyOIDCPatch(current, input)
		return s.validateOIDCCandidate(txCtx, next, input.ClientSecret)
	})
	if snapshotErr != nil {
		return snapshotErr
	}
	if next.Enabled && (forceReadiness || !current.Enabled || current.IssuerURL != next.IssuerURL) {
		readinessErr := s.checkOIDCReadiness(ctx, next.IssuerURL)
		if revisionErr := s.requireCurrentRevision(ctx, flow); revisionErr != nil {
			return revisionErr
		}
		return readinessErr
	}
	return nil
}

func (s *Service) validateOIDCCandidate(
	ctx context.Context,
	settings OIDCSettings,
	secret OptionalSecret,
) error {
	if validationErr := validateOIDC(settings); validationErr != nil {
		return validationErr
	}
	return s.validateRetainedSecret(ctx, settings.Enabled, secret, keyOIDCClientSecret)
}

func (s *Service) validateGoogleCandidate(
	ctx context.Context,
	settings GoogleSettings,
	secret OptionalSecret,
) error {
	if validationErr := validateGoogle(settings); validationErr != nil {
		return validationErr
	}
	return s.validateRetainedSecret(ctx, settings.Enabled, secret, keyGoogleClientSecret)
}

func (s *Service) validateGitHubCandidate(
	ctx context.Context,
	settings GitHubSettings,
	secret OptionalSecret,
) error {
	if validationErr := validateGitHub(settings); validationErr != nil {
		return validationErr
	}
	return s.validateRetainedSecret(ctx, settings.Enabled, secret, keyGitHubClientSecret)
}

func (s *Service) validateRetainedSecret(
	ctx context.Context,
	enabled bool,
	secret OptionalSecret,
	key string,
) error {
	if !enabled || (secret.Present && secret.Value != nil) {
		return nil
	}
	rows, loadErr := s.settingsByKeys(ctx, []string{key})
	if loadErr != nil {
		return loadErr
	}
	value, decryptErr := s.decryptSecret(rows, key)
	if decryptErr != nil || value == "" {
		return invalidField("clientSecret", "clientSecret must be replaced before enabling this provider", nil)
	}
	return nil
}

func (s *Service) ValidateOIDC(ctx context.Context, input ValidateOIDCInput) error {
	ctx, flow := s.startSettingsFlow(ctx, "validate", ResourceOIDC, input.CurrentUserID, input.ExpectedRevision)
	defer flow.end()
	if authorizeErr := flow.authorize(ctx); authorizeErr != nil {
		return authorizeErr
	}
	if secretErr := validateSecretPatch(input.ClientSecret, "clientSecret"); secretErr != nil {
		return secretErr
	}
	normalizeOIDCInput(&input)
	return s.prepareOIDCUpdate(ctx, flow, input, true)
}

func (s *Service) ValidateGoogle(ctx context.Context, input ValidateGoogleInput) error {
	ctx, flow := s.startSettingsFlow(ctx, "validate", ResourceGoogle, input.CurrentUserID, input.ExpectedRevision)
	defer flow.end()
	if authorizeErr := flow.authorize(ctx); authorizeErr != nil {
		return authorizeErr
	}
	if secretErr := validateSecretPatch(input.ClientSecret, "clientSecret"); secretErr != nil {
		return secretErr
	}
	if normalizeErr := normalizeGoogleInput(&input); normalizeErr != nil {
		return normalizeErr
	}
	return validateProviderSnapshot(ctx, s, flow, providerValidationOperations[GoogleSettings]{
		load:  s.effectiveGoogleView,
		apply: func(current GoogleSettings) GoogleSettings { return applyGooglePatch(current, input) },
		validate: func(txCtx context.Context, next GoogleSettings) error {
			return s.validateGoogleCandidate(txCtx, next, input.ClientSecret)
		},
	})
}

func (s *Service) ValidateGitHub(ctx context.Context, input ValidateGitHubInput) error {
	ctx, flow := s.startSettingsFlow(ctx, "validate", ResourceGitHub, input.CurrentUserID, input.ExpectedRevision)
	defer flow.end()
	if authorizeErr := flow.authorize(ctx); authorizeErr != nil {
		return authorizeErr
	}
	if secretErr := validateSecretPatch(input.ClientSecret, "clientSecret"); secretErr != nil {
		return secretErr
	}
	normalizeGitHubInput(&input)
	return validateProviderSnapshot(ctx, s, flow, providerValidationOperations[GitHubSettings]{
		load:  s.effectiveGitHubView,
		apply: func(current GitHubSettings) GitHubSettings { return applyGitHubPatch(current, input) },
		validate: func(txCtx context.Context, next GitHubSettings) error {
			return s.validateGitHubCandidate(txCtx, next, input.ClientSecret)
		},
	})
}

type providerValidationOperations[T any] struct {
	load     func(context.Context) (T, error)
	apply    func(T) T
	validate func(context.Context, T) error
}

func validateProviderSnapshot[T any](
	ctx context.Context,
	service *Service,
	flow settingsFlow,
	operations providerValidationOperations[T],
) error {
	return service.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		revision, lockErr := service.repo.LockSystemSettingRevision(txCtx, string(flow.resource))
		if lockErr != nil {
			return lockErr
		}
		if revisionErr := flow.requireRevision(revision); revisionErr != nil {
			return revisionErr
		}
		current, loadErr := operations.load(txCtx)
		if loadErr != nil {
			return loadErr
		}
		return operations.validate(txCtx, operations.apply(current))
	})
}

func (s *Service) requireCurrentRevision(ctx context.Context, flow settingsFlow) error {
	current, err := s.revision(ctx, flow.resource)
	if err != nil {
		return err
	}
	return flow.requireRevision(current)
}

func (s *Service) checkOIDCReadiness(ctx context.Context, issuerURL string) error {
	if s.readiness == nil {
		return ErrProviderUnavailable
	}
	checkCtx, cancel := context.WithTimeout(ctx, s.readinessTimeout)
	defer cancel()
	if err := s.readiness.Check(checkCtx, issuerURL); err != nil {
		var metadataErr interface{ InvalidOIDCMetadata() }
		if errors.As(err, &metadataErr) {
			return invalidField("issuerUrl", "issuer discovery metadata is invalid or does not match issuerUrl", issuerURL)
		}
		return errors.Join(ErrProviderUnavailable, fmt.Errorf("OIDC discovery unavailable: %w", err))
	}
	return nil
}

func normalizeOIDCInput(input *UpdateOIDCInput) {
	input.IssuerURL = normalizeStringPointer(input.IssuerURL)
	input.ClientID = normalizeStringPointer(input.ClientID)
	input.DisplayName = normalizeStringPointer(input.DisplayName)
}

func normalizeGoogleInput(input *UpdateGoogleInput) error {
	input.ClientID = normalizeStringPointer(input.ClientID)
	input.DisplayName = normalizeStringPointer(input.DisplayName)
	if input.AllowedDomains == nil {
		return nil
	}
	domains, err := normalizeGoogleDomains(*input.AllowedDomains)
	if err != nil {
		return invalidField("allowedDomains", err.Error(), *input.AllowedDomains)
	}
	input.AllowedDomains = &domains
	return nil
}

func normalizeGitHubInput(input *UpdateGitHubInput) {
	input.ClientID = normalizeStringPointer(input.ClientID)
	input.DisplayName = normalizeStringPointer(input.DisplayName)
}

func applyOIDCPatch(current OIDCSettings, input UpdateOIDCInput) OIDCSettings {
	next := current
	if input.Enabled != nil {
		next.Enabled = *input.Enabled
	}
	if input.IssuerURL != nil {
		next.IssuerURL = *input.IssuerURL
	}
	if input.ClientID != nil {
		next.ClientID = *input.ClientID
	}
	if input.ClientSecret.Present {
		next.ClientSecretSet = input.ClientSecret.Value != nil
	}
	if input.DisplayName != nil {
		next.DisplayName = *input.DisplayName
	}
	if input.JITEnabled != nil {
		next.JITEnabled = *input.JITEnabled
	}
	return next
}

func applyGooglePatch(current GoogleSettings, input UpdateGoogleInput) GoogleSettings {
	next := current
	if input.Enabled != nil {
		next.Enabled = *input.Enabled
	}
	if input.ClientID != nil {
		next.ClientID = *input.ClientID
	}
	if input.ClientSecret.Present {
		next.ClientSecretSet = input.ClientSecret.Value != nil
	}
	if input.DisplayName != nil {
		next.DisplayName = *input.DisplayName
	}
	if input.JITEnabled != nil {
		next.JITEnabled = *input.JITEnabled
	}
	if input.AllowedDomains != nil {
		next.AllowedDomains = cloneStrings(*input.AllowedDomains)
	}
	return next
}

func applyGitHubPatch(current GitHubSettings, input UpdateGitHubInput) GitHubSettings {
	next := current
	if input.Enabled != nil {
		next.Enabled = *input.Enabled
	}
	if input.ClientID != nil {
		next.ClientID = *input.ClientID
	}
	if input.ClientSecret.Present {
		next.ClientSecretSet = input.ClientSecret.Value != nil
	}
	if input.DisplayName != nil {
		next.DisplayName = *input.DisplayName
	}
	if input.JITEnabled != nil {
		next.JITEnabled = *input.JITEnabled
	}
	if input.AllowSignup != nil {
		next.AllowSignup = *input.AllowSignup
	}
	return next
}

func oidcPatchIsNoop(current, next OIDCSettings, secret OptionalSecret) bool {
	return reflect.DeepEqual(current, next) && !secretReplacement(secret, current.ClientSecretSet)
}

func googlePatchIsNoop(current, next GoogleSettings, secret OptionalSecret) bool {
	return reflect.DeepEqual(current, next) && !secretReplacement(secret, current.ClientSecretSet)
}

func gitHubPatchIsNoop(current, next GitHubSettings, secret OptionalSecret) bool {
	return reflect.DeepEqual(current, next) && !secretReplacement(secret, current.ClientSecretSet)
}

func secretReplacement(secret OptionalSecret, currentlySet bool) bool {
	if !secret.Present {
		return false
	}
	return secret.Value != nil || currentlySet
}

func (s *Service) persistOIDC(
	ctx context.Context,
	current, next OIDCSettings,
	secret OptionalSecret,
	actor *string,
) error {
	currentStored := storedOIDCSettings{
		Enabled: current.Enabled, IssuerURL: current.IssuerURL, ClientID: current.ClientID,
		DisplayName: current.DisplayName, JITEnabled: current.JITEnabled,
	}
	nextStored := storedOIDCSettings{
		Enabled: next.Enabled, IssuerURL: next.IssuerURL, ClientID: next.ClientID,
		DisplayName: next.DisplayName, JITEnabled: next.JITEnabled,
	}
	return s.persistProvider(
		ctx,
		keyOIDCSettings,
		keyOIDCClientSecret,
		currentStored,
		nextStored,
		current.ClientSecretSet,
		secret,
		actor,
	)
}

func (s *Service) persistGoogle(
	ctx context.Context,
	current, next GoogleSettings,
	secret OptionalSecret,
	actor *string,
) error {
	currentStored := storedGoogleSettings{
		Enabled: current.Enabled, ClientID: current.ClientID, DisplayName: current.DisplayName,
		JITEnabled: current.JITEnabled, AllowedDomains: cloneStrings(current.AllowedDomains),
	}
	nextStored := storedGoogleSettings{
		Enabled: next.Enabled, ClientID: next.ClientID, DisplayName: next.DisplayName,
		JITEnabled: next.JITEnabled, AllowedDomains: cloneStrings(next.AllowedDomains),
	}
	return s.persistProvider(
		ctx,
		keyGoogleSettings,
		keyGoogleClientSecret,
		currentStored,
		nextStored,
		current.ClientSecretSet,
		secret,
		actor,
	)
}

func (s *Service) persistGitHub(
	ctx context.Context,
	current, next GitHubSettings,
	secret OptionalSecret,
	actor *string,
) error {
	currentStored := storedGitHubSettings{
		Enabled: current.Enabled, ClientID: current.ClientID, DisplayName: current.DisplayName,
		JITEnabled: current.JITEnabled, AllowSignup: current.AllowSignup,
	}
	nextStored := storedGitHubSettings{
		Enabled: next.Enabled, ClientID: next.ClientID, DisplayName: next.DisplayName,
		JITEnabled: next.JITEnabled, AllowSignup: next.AllowSignup,
	}
	return s.persistProvider(
		ctx,
		keyGitHubSettings,
		keyGitHubClientSecret,
		currentStored,
		nextStored,
		current.ClientSecretSet,
		secret,
		actor,
	)
}

func (s *Service) persistProvider(
	ctx context.Context,
	publicKey string,
	secretKey string,
	currentPublic any,
	nextPublic any,
	currentSecretSet bool,
	secret OptionalSecret,
	actor *string,
) error {
	if !reflect.DeepEqual(currentPublic, nextPublic) {
		if publicErr := s.upsertPublic(ctx, publicKey, nextPublic, actor); publicErr != nil {
			return publicErr
		}
	}
	return s.persistProviderSecret(ctx, secretKey, currentSecretSet, secret, actor)
}

func (s *Service) persistProviderSecret(
	ctx context.Context,
	key string,
	currentlySet bool,
	secret OptionalSecret,
	actor *string,
) error {
	if !secret.Present {
		return nil
	}
	if secret.Value == nil {
		if !currentlySet {
			return nil
		}
		return s.deleteSecret(ctx, key, actor)
	}
	return s.upsertSecret(ctx, key, *secret.Value, actor)
}
