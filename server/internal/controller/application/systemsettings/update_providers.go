package systemsettings

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

func (s *Service) UpdateOIDC(ctx context.Context, input UpdateOIDCInput) (OIDCSettings, error) {
	ctx, flow := s.startSettingsFlow(ctx, "update", ResourceOIDC, input.CurrentUserID)
	defer flow.end()
	if authorizeErr := flow.authorize(ctx); authorizeErr != nil {
		return OIDCSettings{}, authorizeErr
	}
	if secretErr := validateSecretPatch(input.ClientSecret, "clientSecret"); secretErr != nil {
		return OIDCSettings{}, secretErr
	}
	normalizeOIDCInput(&input)

	var result OIDCSettings
	txErr := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		if lockErr := s.repo.LockSystemSettingsResource(txCtx, string(ResourceOIDC)); lockErr != nil {
			return lockErr
		}
		current, loadErr := s.effectiveOIDCView(txCtx)
		if loadErr != nil {
			return loadErr
		}
		next := applyOIDCPatch(current, input)
		if validationErr := s.validateOIDCCandidate(txCtx, next, input.ClientSecret); validationErr != nil {
			return validationErr
		}
		if next.Enabled && (!current.Enabled || current.IssuerURL != next.IssuerURL) {
			if readinessErr := s.checkOIDCReadiness(txCtx, next.IssuerURL); readinessErr != nil {
				return readinessErr
			}
		}
		if oidcPatchIsNoop(current, next, input.ClientSecret) {
			result = current
			return nil
		}
		if persistErr := s.persistOIDC(txCtx, current, next, input.ClientSecret, &input.CurrentUserID); persistErr != nil {
			return persistErr
		}
		result = next
		return nil
	})
	return result, txErr
}

func (s *Service) UpdateGoogle(ctx context.Context, input UpdateGoogleInput) (GoogleSettings, error) {
	ctx, flow := s.startSettingsFlow(ctx, "update", ResourceGoogle, input.CurrentUserID)
	defer flow.end()
	if authorizeErr := flow.authorize(ctx); authorizeErr != nil {
		return GoogleSettings{}, authorizeErr
	}
	if secretErr := validateSecretPatch(input.ClientSecret, "clientSecret"); secretErr != nil {
		return GoogleSettings{}, secretErr
	}
	if normalizeErr := normalizeGoogleInput(&input); normalizeErr != nil {
		return GoogleSettings{}, normalizeErr
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

func (s *Service) UpdateGitHub(ctx context.Context, input UpdateGitHubInput) (GitHubSettings, error) {
	ctx, flow := s.startSettingsFlow(ctx, "update", ResourceGitHub, input.CurrentUserID)
	defer flow.end()
	if authorizeErr := flow.authorize(ctx); authorizeErr != nil {
		return GitHubSettings{}, authorizeErr
	}
	if secretErr := validateSecretPatch(input.ClientSecret, "clientSecret"); secretErr != nil {
		return GitHubSettings{}, secretErr
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
) (T, error) {
	var result T
	txErr := service.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		if lockErr := service.repo.LockSystemSettingsResource(txCtx, string(flow.resource)); lockErr != nil {
			return lockErr
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
			result = current
			return nil
		}
		if persistErr := operations.persist(txCtx, current, next); persistErr != nil {
			return persistErr
		}
		result = next
		return nil
	})
	return result, txErr
}

func (s *Service) validateOIDCSnapshot(ctx context.Context, input ValidateOIDCInput) error {
	var current OIDCSettings
	var next OIDCSettings
	snapshotErr := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
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
	if next.Enabled {
		return s.checkOIDCReadiness(ctx, next.IssuerURL)
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
	ctx, flow := s.startSettingsFlow(ctx, "validate", ResourceOIDC, input.CurrentUserID)
	defer flow.end()
	if authorizeErr := flow.authorize(ctx); authorizeErr != nil {
		return authorizeErr
	}
	if secretErr := validateSecretPatch(input.ClientSecret, "clientSecret"); secretErr != nil {
		return secretErr
	}
	normalizeOIDCInput(&input)
	return s.validateOIDCSnapshot(ctx, input)
}

func (s *Service) ValidateGoogle(ctx context.Context, input ValidateGoogleInput) error {
	ctx, flow := s.startSettingsFlow(ctx, "validate", ResourceGoogle, input.CurrentUserID)
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
	return validateProviderSnapshot(ctx, s, providerValidationOperations[GoogleSettings]{
		load:  s.effectiveGoogleView,
		apply: func(current GoogleSettings) GoogleSettings { return applyGooglePatch(current, input) },
		validate: func(txCtx context.Context, next GoogleSettings) error {
			return s.validateGoogleCandidate(txCtx, next, input.ClientSecret)
		},
	})
}

func (s *Service) ValidateGitHub(ctx context.Context, input ValidateGitHubInput) error {
	ctx, flow := s.startSettingsFlow(ctx, "validate", ResourceGitHub, input.CurrentUserID)
	defer flow.end()
	if authorizeErr := flow.authorize(ctx); authorizeErr != nil {
		return authorizeErr
	}
	if secretErr := validateSecretPatch(input.ClientSecret, "clientSecret"); secretErr != nil {
		return secretErr
	}
	normalizeGitHubInput(&input)
	return validateProviderSnapshot(ctx, s, providerValidationOperations[GitHubSettings]{
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
	operations providerValidationOperations[T],
) error {
	return service.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		current, loadErr := operations.load(txCtx)
		if loadErr != nil {
			return loadErr
		}
		return operations.validate(txCtx, operations.apply(current))
	})
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
