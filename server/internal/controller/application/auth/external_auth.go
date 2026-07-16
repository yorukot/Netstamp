package auth

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

const (
	ExternalAuthIntentLogin = "login"
	ExternalAuthIntentSudo  = "sudo"
	ExternalAuthIntentLink  = "link"

	// Compatibility aliases for callers using the original generic OIDC API.
	OIDCIntentLogin = ExternalAuthIntentLogin
	OIDCIntentSudo  = ExternalAuthIntentSudo
	OIDCIntentLink  = ExternalAuthIntentLink
)

func (s *Service) ExternalProviderMethods() []ExternalProviderMethod {
	methods := make([]ExternalProviderMethod, 0, len(s.externalProviders))
	for _, provider := range s.externalProviders {
		methods = append(methods, ExternalProviderMethod{
			ID: provider.config.ID, DisplayName: provider.config.DisplayName, SudoCapable: provider.config.SudoCapable,
		})
	}
	order := map[string]int{
		identity.AuthenticationMethodGoogle: 0,
		identity.AuthenticationMethodGitHub: 1,
		identity.AuthenticationMethodOIDC:   2,
	}
	sort.Slice(methods, func(i, j int) bool {
		left, leftKnown := order[methods[i].ID]
		right, rightKnown := order[methods[j].ID]
		if leftKnown && rightKnown && left != right {
			return left < right
		}
		if leftKnown != rightKnown {
			return leftKnown
		}
		return methods[i].ID < methods[j].ID
	})
	return methods
}

func (s *Service) StartExternalAuth(ctx context.Context, input StartExternalAuthInput) (StartExternalAuthResult, error) {
	provider, err := s.prepareExternalAuthStart(ctx, &input)
	if err != nil {
		return StartExternalAuthResult{}, err
	}
	secrets, err := s.generateExternalAuthFlowSecrets(ctx)
	if err != nil {
		return StartExternalAuthResult{}, err
	}
	now := s.now()
	if deleteErr := s.externalAuthRepo.DeleteExpiredExternalAuthFlows(ctx, now); deleteErr != nil {
		return StartExternalAuthResult{}, deleteErr
	}
	expiresAt := now.Add(s.externalAuthConfig.FlowTTL)
	_, err = s.externalAuthRepo.CreateExternalAuthFlow(ctx, identity.ExternalAuthFlow{
		Provider: input.Provider, StateHash: []byte(s.externalAuthTokens.Hash(secrets.state)),
		BrowserTokenHash: []byte(s.externalAuthTokens.Hash(secrets.browserToken)), Nonce: secrets.nonce,
		PKCEVerifier: secrets.pkceVerifier, Intent: input.Intent, SessionID: optionalSessionID(input.SessionID),
		ReturnTo: normalizeReturnTo(input.ReturnTo), CreatedAt: now, ExpiresAt: expiresAt,
	})
	if err != nil {
		return StartExternalAuthResult{}, err
	}
	authorizationURL, err := provider.client.AuthorizationURL(ctx, secrets.state, secrets.nonce, secrets.pkceVerifier, input.Intent)
	if err != nil {
		return StartExternalAuthResult{}, ErrExternalAuthUnavailable
	}
	return StartExternalAuthResult{AuthorizationURL: authorizationURL, BrowserToken: secrets.browserToken, ExpiresAt: expiresAt}, nil
}

type externalAuthFlowSecrets struct {
	state        string
	browserToken string
	nonce        string
	pkceVerifier string
}

func (s *Service) prepareExternalAuthStart(ctx context.Context, input *StartExternalAuthInput) (configuredExternalProvider, error) {
	input.Provider = strings.ToLower(strings.TrimSpace(input.Provider))
	provider, ok := s.externalProviders[input.Provider]
	if !ok || !s.externalAuthAvailable() {
		return configuredExternalProvider{}, ErrExternalAuthUnavailable
	}
	if input.Intent == "" {
		input.Intent = ExternalAuthIntentLogin
	}
	switch input.Intent {
	case ExternalAuthIntentLogin:
		return provider, nil
	case ExternalAuthIntentSudo:
		if !provider.config.SudoCapable {
			return configuredExternalProvider{}, ErrExternalAuthSudoUnsupported
		}
		if input.SessionID == "" || s.recentAuth == nil {
			return configuredExternalProvider{}, ErrSessionInvalid
		}
		return provider, nil
	case ExternalAuthIntentLink:
		if input.SessionID == "" || s.recentAuth == nil {
			return configuredExternalProvider{}, ErrSessionInvalid
		}
		status, err := s.recentAuth.SudoStatus(ctx, input.SessionID)
		if err != nil {
			return configuredExternalProvider{}, err
		}
		if !status.Active {
			return configuredExternalProvider{}, ErrSudoRequired
		}
		return provider, nil
	default:
		return configuredExternalProvider{}, ErrInvalidInput
	}
}

func (s *Service) generateExternalAuthFlowSecrets(ctx context.Context) (externalAuthFlowSecrets, error) {
	values := make([]string, 4)
	for i := range values {
		value, err := s.externalAuthTokens.Generate(ctx)
		if err != nil {
			return externalAuthFlowSecrets{}, err
		}
		values[i] = value
	}
	return externalAuthFlowSecrets{state: values[0], browserToken: values[1], nonce: values[2], pkceVerifier: values[3]}, nil
}

func optionalSessionID(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func (s *Service) CompleteExternalAuth(ctx context.Context, input CompleteExternalAuthInput) (CompleteExternalAuthResult, error) {
	input.Provider = strings.ToLower(strings.TrimSpace(input.Provider))
	provider, ok := s.externalProviders[input.Provider]
	if !ok || !s.externalAuthAvailable() {
		return CompleteExternalAuthResult{}, ErrExternalAuthUnavailable
	}
	if input.Code == "" || input.State == "" || input.BrowserToken == "" {
		return CompleteExternalAuthResult{}, ErrExternalAuthCallbackInvalid
	}
	now := s.now()
	flow, err := s.externalAuthRepo.ConsumeExternalAuthFlow(
		ctx, input.Provider, []byte(s.externalAuthTokens.Hash(input.State)), []byte(s.externalAuthTokens.Hash(input.BrowserToken)), now,
	)
	if err != nil || flow.Provider != input.Provider {
		return CompleteExternalAuthResult{}, ErrExternalAuthCallbackInvalid
	}
	claims, err := provider.client.Exchange(ctx, input.Code, flow.PKCEVerifier, flow.Nonce)
	if err != nil || claims.Issuer == "" || claims.Subject == "" {
		return CompleteExternalAuthResult{}, ErrExternalAuthCallbackInvalid
	}
	return s.completeExternalAuthIntent(ctx, provider, flow, claims, input.UserAgent, now)
}

func (s *Service) completeExternalAuthIntent(
	ctx context.Context,
	provider configuredExternalProvider,
	flow identity.ExternalAuthFlow,
	claims ExternalIdentityClaims,
	userAgent string,
	now time.Time,
) (CompleteExternalAuthResult, error) {
	result := CompleteExternalAuthResult{Intent: flow.Intent, ReturnTo: flow.ReturnTo}
	switch flow.Intent {
	case ExternalAuthIntentLogin:
		access, loginErr := s.completeExternalAuthLogin(ctx, provider, claims, userAgent, now)
		if loginErr != nil {
			return result, loginErr
		}
		result.Access = &access
		return result, nil
	case ExternalAuthIntentSudo:
		return result, s.completeExternalAuthSudo(ctx, provider, flow, claims, now)
	case ExternalAuthIntentLink:
		return result, s.completeExternalAuthLink(ctx, provider, flow, claims, now)
	default:
		return result, ErrExternalAuthCallbackInvalid
	}
}

func (s *Service) completeExternalAuthSudo(ctx context.Context, provider configuredExternalProvider, flow identity.ExternalAuthFlow, claims ExternalIdentityClaims, now time.Time) error {
	if !provider.config.SudoCapable {
		return ErrExternalAuthSudoUnsupported
	}
	session, authenticatedAt, err := s.externalReauthenticationSession(ctx, flow, claims.AuthTime, now)
	if err != nil {
		return err
	}
	linked, err := s.externalAuthRepo.GetUserIdentityByIssuerSubject(ctx, provider.config.ID, claims.Issuer, claims.Subject)
	if err != nil || linked.UserID != session.UserID {
		return ErrIdentityConflict
	}
	identityID := linked.ID
	return s.recentAuth.ElevateSession(ctx, session.ID, provider.config.ID, &identityID, authenticatedAt)
}

func (s *Service) completeExternalAuthLink(ctx context.Context, provider configuredExternalProvider, flow identity.ExternalAuthFlow, claims ExternalIdentityClaims, now time.Time) error {
	session, err := s.externalLinkSession(ctx, flow)
	if err != nil {
		return err
	}
	_, err = s.externalAuthRepo.GetUserIdentityByIssuerSubject(ctx, provider.config.ID, claims.Issuer, claims.Subject)
	switch {
	case err == nil:
		return ErrIdentityConflict
	case !errors.Is(err, identity.ErrIdentityNotFound):
		return err
	}
	linked, err := s.externalAuthRepo.CreateUserIdentity(ctx, identity.UserIdentity{
		UserID: session.UserID, Provider: provider.config.ID, Issuer: claims.Issuer, Subject: claims.Subject,
		Email: optionalString(claims.Email), EmailVerified: claims.EmailVerified, DisplayName: optionalString(claims.DisplayName),
		Username: optionalString(claims.Username), AvatarURL: optionalString(claims.AvatarURL), CreatedAt: now, LastLoginAt: &now,
	})
	if err != nil {
		return ErrIdentityConflict
	}
	_ = linked
	return nil
}

func (s *Service) externalLinkSession(ctx context.Context, flow identity.ExternalAuthFlow) (identity.AuthSession, error) {
	if flow.SessionID == nil || s.recentAuth == nil {
		return identity.AuthSession{}, ErrExternalAuthCallbackInvalid
	}
	status, err := s.recentAuth.SudoStatus(ctx, *flow.SessionID)
	if err != nil {
		return identity.AuthSession{}, err
	}
	if !status.Active {
		return identity.AuthSession{}, ErrSudoRequired
	}
	return s.recentAuth.GetSession(ctx, *flow.SessionID)
}

func (s *Service) externalReauthenticationSession(ctx context.Context, flow identity.ExternalAuthFlow, authTime, now time.Time) (identity.AuthSession, time.Time, error) {
	if flow.SessionID == nil || s.recentAuth == nil {
		return identity.AuthSession{}, time.Time{}, ErrExternalAuthCallbackInvalid
	}
	session, err := s.recentAuth.GetSession(ctx, *flow.SessionID)
	if err != nil {
		return identity.AuthSession{}, time.Time{}, err
	}
	authenticatedAt, ok := recentExternalAuthenticationTime(authTime, flow.CreatedAt, session.CreatedAt, now, s.externalAuthConfig.AuthTimeSkew)
	if !ok {
		return identity.AuthSession{}, time.Time{}, ErrExternalAuthCallbackInvalid
	}
	return session, authenticatedAt, nil
}

func (s *Service) externalAuthAvailable() bool {
	return s.externalAuthRepo != nil && s.externalAuthTokens != nil && len(s.externalProviders) > 0
}

func recentExternalAuthenticationTime(authTime, flowCreatedAt, sessionCreatedAt, now time.Time, skew time.Duration) (time.Time, bool) {
	if authTime.IsZero() || authTime.Before(flowCreatedAt.Add(-skew)) || authTime.After(now.Add(skew)) {
		return time.Time{}, false
	}
	authTime = authTime.UTC()
	if !sessionCreatedAt.IsZero() && authTime.Before(sessionCreatedAt) {
		authTime = sessionCreatedAt.UTC()
	}
	if authTime.After(now) {
		authTime = now.UTC()
	}
	return authTime, true
}

func (s *Service) completeExternalAuthLogin(ctx context.Context, provider configuredExternalProvider, claims ExternalIdentityClaims, userAgent string, now time.Time) (AuthAccessResult, error) {
	linked, err := s.externalAuthRepo.GetUserIdentityByIssuerSubject(ctx, provider.config.ID, claims.Issuer, claims.Subject)
	var user identity.User
	switch {
	case errors.Is(err, identity.ErrIdentityNotFound):
		user, linked, err = s.provisionExternalAuthUser(ctx, provider, claims, now)
	case err != nil:
		return AuthAccessResult{}, err
	default:
		user, err = s.users.GetUserByID(ctx, linked.UserID)
	}
	if err != nil {
		return AuthAccessResult{}, err
	}
	if user.DisabledAt != nil {
		return AuthAccessResult{}, ErrAccountDisabled
	}
	linked.Email = optionalString(claims.Email)
	linked.EmailVerified = claims.EmailVerified
	linked.DisplayName = optionalString(claims.DisplayName)
	linked.Username = optionalString(claims.Username)
	linked.AvatarURL = optionalString(claims.AvatarURL)
	if _, err := s.externalAuthRepo.TouchUserIdentityLogin(ctx, linked, now); err != nil {
		return AuthAccessResult{}, err
	}
	identityID := linked.ID
	return s.createAccessResultWithMethod(ctx, user, userAgent, provider.config.ID, false, &identityID)
}

func (s *Service) provisionExternalAuthUser(ctx context.Context, provider configuredExternalProvider, claims ExternalIdentityClaims, now time.Time) (identity.User, identity.UserIdentity, error) {
	if !provider.config.JITEnabled {
		return identity.User{}, identity.UserIdentity{}, ErrJITProvisioningDisabled
	}
	if !claims.EmailVerified {
		return identity.User{}, identity.UserIdentity{}, ErrExternalAuthCallbackInvalid
	}
	email, err := identity.VNUserEmail(claims.Email)
	if err != nil {
		return identity.User{}, identity.UserIdentity{}, ErrExternalAuthCallbackInvalid
	}
	_, lookupErr := s.users.GetUserByEmail(ctx, email)
	switch {
	case lookupErr == nil:
		return identity.User{}, identity.UserIdentity{}, ErrIdentityConflict
	case !errors.Is(lookupErr, identity.ErrUserNotFound):
		return identity.User{}, identity.UserIdentity{}, lookupErr
	}
	displayName := strings.TrimSpace(claims.DisplayName)
	if displayName == "" {
		displayName = strings.Split(email, "@")[0]
	}
	displayName, err = identity.VNUserDisplayName(displayName)
	if err != nil {
		return identity.User{}, identity.UserIdentity{}, ErrExternalAuthCallbackInvalid
	}
	var user identity.User
	var linked identity.UserIdentity
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		created, createdIdentity, createErr := s.externalAuthRepo.CreateExternalAuthUser(ctx, email, displayName, identity.UserIdentity{
			Provider: provider.config.ID, Issuer: claims.Issuer, Subject: claims.Subject, Email: optionalString(email),
			EmailVerified: true, DisplayName: optionalString(displayName), Username: optionalString(claims.Username), AvatarURL: optionalString(claims.AvatarURL),
		}, now)
		if createErr != nil {
			return createErr
		}
		created, createErr = s.applyFirstSystemAdminGrant(ctx, created, now)
		user, linked = created, createdIdentity
		return createErr
	})
	if errors.Is(err, identity.ErrEmailAlreadyExists) || errors.Is(err, identity.ErrIdentityConflict) {
		return identity.User{}, identity.UserIdentity{}, ErrIdentityConflict
	}
	return user, linked, err
}

func normalizeReturnTo(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") || strings.ContainsAny(value, "\\\r\n") {
		return "/"
	}
	return value
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
