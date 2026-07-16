package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

const (
	OIDCIntentLogin = "login"
	OIDCIntentSudo  = "sudo"
	OIDCIntentLink  = "link"
)

func (s *Service) StartOIDC(ctx context.Context, input StartOIDCInput) (StartOIDCResult, error) {
	if err := s.prepareOIDCStart(ctx, &input); err != nil {
		return StartOIDCResult{}, err
	}
	secrets, err := s.generateOIDCFlowSecrets(ctx)
	if err != nil {
		return StartOIDCResult{}, err
	}
	now := s.now()
	err = s.oidcRepo.DeleteExpiredOIDCAuthFlows(ctx, now)
	if err != nil {
		return StartOIDCResult{}, err
	}
	_, err = s.oidcRepo.CreateOIDCAuthFlow(ctx, identity.OIDCAuthFlow{
		StateHash: []byte(s.oidcTokens.Hash(secrets.state)), BrowserTokenHash: []byte(s.oidcTokens.Hash(secrets.browserToken)),
		Nonce: secrets.nonce, PKCEVerifier: secrets.pkceVerifier, Intent: input.Intent, SessionID: optionalSessionID(input.SessionID),
		ReturnTo: normalizeReturnTo(input.ReturnTo), CreatedAt: now, ExpiresAt: now.Add(s.oidcConfig.FlowTTL),
	})
	if err != nil {
		return StartOIDCResult{}, err
	}
	url, err := s.oidcClient.AuthorizationURL(ctx, secrets.state, secrets.nonce, secrets.pkceVerifier, input.Intent != OIDCIntentLogin)
	if err != nil {
		return StartOIDCResult{}, ErrOIDCUnavailable
	}
	return StartOIDCResult{AuthorizationURL: url, BrowserToken: secrets.browserToken, ExpiresAt: now.Add(s.oidcConfig.FlowTTL)}, nil
}

type oidcFlowSecrets struct {
	state        string
	browserToken string
	nonce        string
	pkceVerifier string
}

func (s *Service) prepareOIDCStart(ctx context.Context, input *StartOIDCInput) error {
	if !s.oidcAvailable() {
		return ErrOIDCUnavailable
	}
	if input.Intent == "" {
		input.Intent = OIDCIntentLogin
	}
	switch input.Intent {
	case OIDCIntentLogin:
		return nil
	case OIDCIntentSudo:
		if input.SessionID == "" || s.recentAuth == nil {
			return ErrSessionInvalid
		}
		return nil
	case OIDCIntentLink:
		if input.SessionID == "" || s.recentAuth == nil {
			return ErrSessionInvalid
		}
		status, err := s.recentAuth.SudoStatus(ctx, input.SessionID)
		if err != nil {
			return err
		}
		if !status.Active {
			return ErrSudoRequired
		}
		return nil
	default:
		return ErrInvalidInput
	}
}

func (s *Service) generateOIDCFlowSecrets(ctx context.Context) (oidcFlowSecrets, error) {
	values := make([]string, 4)
	for i := range values {
		value, err := s.oidcTokens.Generate(ctx)
		if err != nil {
			return oidcFlowSecrets{}, err
		}
		values[i] = value
	}
	return oidcFlowSecrets{state: values[0], browserToken: values[1], nonce: values[2], pkceVerifier: values[3]}, nil
}

func optionalSessionID(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func (s *Service) CompleteOIDC(ctx context.Context, input CompleteOIDCInput) (CompleteOIDCResult, error) {
	if !s.oidcAvailable() {
		return CompleteOIDCResult{}, ErrOIDCUnavailable
	}
	if input.Code == "" || input.State == "" || input.BrowserToken == "" {
		return CompleteOIDCResult{}, ErrOIDCCallbackInvalid
	}
	now := s.now()
	flow, err := s.oidcRepo.ConsumeOIDCAuthFlow(ctx, []byte(s.oidcTokens.Hash(input.State)), []byte(s.oidcTokens.Hash(input.BrowserToken)), now)
	if err != nil {
		return CompleteOIDCResult{}, ErrOIDCCallbackInvalid
	}
	claims, err := s.oidcClient.Exchange(ctx, input.Code, flow.PKCEVerifier, flow.Nonce)
	if err != nil || claims.Issuer == "" || claims.Subject == "" {
		return CompleteOIDCResult{}, ErrOIDCCallbackInvalid
	}
	result := CompleteOIDCResult{Intent: flow.Intent, ReturnTo: flow.ReturnTo}
	switch flow.Intent {
	case OIDCIntentLogin:
		access, loginErr := s.completeOIDCLogin(ctx, claims, input.UserAgent, now)
		if loginErr != nil {
			return CompleteOIDCResult{}, loginErr
		}
		result.Access = &access
		return result, nil
	case OIDCIntentSudo:
		return result, s.completeOIDCSudo(ctx, flow, claims, now)
	case OIDCIntentLink:
		return result, s.completeOIDCLink(ctx, flow, claims, now)
	default:
		return CompleteOIDCResult{}, ErrOIDCCallbackInvalid
	}
}

func (s *Service) completeOIDCSudo(ctx context.Context, flow identity.OIDCAuthFlow, claims OIDCClaims, now time.Time) error {
	session, authenticatedAt, err := s.oidcReauthenticationSession(ctx, flow, claims.AuthTime, now)
	if err != nil {
		return err
	}
	linked, err := s.oidcRepo.GetUserIdentityByIssuerSubject(ctx, claims.Issuer, claims.Subject)
	if err != nil || linked.UserID != session.UserID {
		return ErrIdentityConflict
	}
	identityID := linked.ID
	return s.recentAuth.ElevateSession(ctx, session.ID, identity.AuthenticationMethodOIDC, &identityID, authenticatedAt)
}

func (s *Service) completeOIDCLink(ctx context.Context, flow identity.OIDCAuthFlow, claims OIDCClaims, now time.Time) error {
	session, authenticatedAt, err := s.oidcReauthenticationSession(ctx, flow, claims.AuthTime, now)
	if err != nil {
		return err
	}
	_, err = s.oidcRepo.GetUserIdentityByIssuerSubject(ctx, claims.Issuer, claims.Subject)
	switch {
	case err == nil:
		return ErrIdentityConflict
	case !errors.Is(err, identity.ErrIdentityNotFound):
		return err
	}
	email, displayName := optionalString(claims.Email), optionalString(claims.DisplayName)
	linked, err := s.oidcRepo.CreateUserIdentity(ctx, identity.UserIdentity{
		UserID: session.UserID, Provider: identity.AuthenticationMethodOIDC, Issuer: claims.Issuer, Subject: claims.Subject,
		Email: email, EmailVerified: claims.EmailVerified, DisplayName: displayName, CreatedAt: now, LastLoginAt: &now,
	})
	if err != nil {
		return ErrIdentityConflict
	}
	identityID := linked.ID
	return s.recentAuth.ElevateSession(ctx, session.ID, identity.AuthenticationMethodOIDC, &identityID, authenticatedAt)
}

func (s *Service) oidcReauthenticationSession(ctx context.Context, flow identity.OIDCAuthFlow, authTime, now time.Time) (identity.AuthSession, time.Time, error) {
	if flow.SessionID == nil || s.recentAuth == nil {
		return identity.AuthSession{}, time.Time{}, ErrOIDCCallbackInvalid
	}
	session, err := s.recentAuth.GetSession(ctx, *flow.SessionID)
	if err != nil {
		return identity.AuthSession{}, time.Time{}, err
	}
	authenticatedAt, ok := recentOIDCAuthenticationTime(authTime, flow.CreatedAt, session.CreatedAt, now, s.oidcConfig.AuthTimeSkew)
	if !ok {
		return identity.AuthSession{}, time.Time{}, ErrOIDCCallbackInvalid
	}
	return session, authenticatedAt, nil
}

func (s *Service) oidcAvailable() bool {
	return s.oidcConfig.Enabled && s.oidcRepo != nil && s.oidcClient != nil && s.oidcTokens != nil
}

func recentOIDCAuthenticationTime(authTime, flowCreatedAt, sessionCreatedAt, now time.Time, skew time.Duration) (time.Time, bool) {
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

func (s *Service) completeOIDCLogin(ctx context.Context, claims OIDCClaims, userAgent string, now time.Time) (AuthAccessResult, error) {
	linked, err := s.oidcRepo.GetUserIdentityByIssuerSubject(ctx, claims.Issuer, claims.Subject)
	var user identity.User
	switch {
	case errors.Is(err, identity.ErrIdentityNotFound):
		user, linked, err = s.provisionOIDCUser(ctx, claims, now)
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
	linked.Email, linked.DisplayName = optionalString(claims.Email), optionalString(claims.DisplayName)
	linked.EmailVerified = claims.EmailVerified
	if _, err := s.oidcRepo.TouchUserIdentityLogin(ctx, linked, now); err != nil {
		return AuthAccessResult{}, err
	}
	identityID := linked.ID
	return s.createAccessResultWithMethod(ctx, user, userAgent, identity.AuthenticationMethodOIDC, &identityID)
}

func (s *Service) provisionOIDCUser(ctx context.Context, claims OIDCClaims, now time.Time) (identity.User, identity.UserIdentity, error) {
	if !s.oidcConfig.JITEnabled {
		return identity.User{}, identity.UserIdentity{}, ErrJITProvisioningDisabled
	}
	if !claims.EmailVerified {
		return identity.User{}, identity.UserIdentity{}, ErrOIDCCallbackInvalid
	}
	email, err := identity.VNUserEmail(claims.Email)
	if err != nil {
		return identity.User{}, identity.UserIdentity{}, ErrOIDCCallbackInvalid
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
		return identity.User{}, identity.UserIdentity{}, ErrOIDCCallbackInvalid
	}
	var user identity.User
	var linked identity.UserIdentity
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		created, createdIdentity, createErr := s.oidcRepo.CreateOIDCUser(ctx, email, displayName, claims.Issuer, claims.Subject, now)
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
