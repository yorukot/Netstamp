package auth

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

type Service struct {
	users       UserRepository
	hasher      PasswordHasher
	tokens      TokenIssuer
	events      SecurityEventRecorder
	resets      PasswordResetRepository
	resetTokens PasswordResetTokenManager
	resetMailer PasswordResetMailer
	resetConfig PasswordResetConfig
	now         func() time.Time
}

func NewService(users UserRepository, hasher PasswordHasher, tokens TokenIssuer, events SecurityEventRecorder) *Service {
	return &Service{
		users:       users,
		hasher:      hasher,
		tokens:      tokens,
		events:      events,
		resetConfig: PasswordResetConfig{TokenTTL: 30 * time.Minute},
		now:         func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) ConfigurePasswordReset(resets PasswordResetRepository, tokens PasswordResetTokenManager, mailer PasswordResetMailer, cfg PasswordResetConfig) {
	s.resets = resets
	s.resetTokens = tokens
	s.resetMailer = mailer
	if cfg.TokenTTL <= 0 {
		cfg.TokenTTL = 30 * time.Minute
	}
	s.resetConfig = cfg
}

// Register is the service entry for the register action
func (s *Service) Register(ctx context.Context, input RegisterInput) (AuthAccessResult, error) {
	ctx, flow := s.startAuthFlow(ctx, "auth.register", AuthActionRegister, input.Email)
	defer flow.end()

	input, err := normalizeRegisterInput(input)
	if err != nil {
		return AuthAccessResult{}, flow.businessFailure(AuthEventRegisterFailure, AuthReasonInvalidInput, err)
	}

	passwordHash, err := s.hashPassword(ctx, input.Password)
	if err != nil {
		return AuthAccessResult{}, flow.technicalFailure(AuthEventRegisterFailure, AuthReasonPasswordHashFailed, err)
	}

	user, err := s.createUser(ctx, identity.User{
		Email:        input.Email,
		DisplayName:  input.DisplayName,
		PasswordHash: passwordHash,
	})
	if errors.Is(err, identity.ErrEmailAlreadyExists) {
		return AuthAccessResult{}, flow.businessFailure(AuthEventRegisterFailure, AuthReasonEmailAlreadyExists, err)
	}
	if err != nil {
		return AuthAccessResult{}, flow.technicalFailure(AuthEventRegisterFailure, AuthReasonUserCreateFailed, err)
	}
	flow.setUser(user)

	result, err := s.issueAccessResult(ctx, user)
	if err != nil {
		return AuthAccessResult{}, flow.technicalFailure(AuthEventTokenIssueFailure, AuthReasonAccessTokenIssueFail, err)
	}

	flow.success(AuthEventRegisterSuccess)

	return result, nil
}

// Login is the enrty for the login action
func (s *Service) Login(ctx context.Context, input LoginInput) (AuthAccessResult, error) {
	ctx, flow := s.startAuthFlow(ctx, "auth.login", AuthActionLogin, input.Email)
	defer flow.end()

	input, err := normalizeLoginInput(input)
	if err != nil {
		return AuthAccessResult{}, flow.businessFailure(AuthEventLoginFailure, AuthReasonInvalidInput, err)
	}

	user, err := s.getUserByEmail(ctx, input.Email)
	if errors.Is(err, identity.ErrUserNotFound) {
		return AuthAccessResult{}, flow.businessFailure(AuthEventLoginFailure, AuthReasonCredentialsInvalid, ErrCredentialsInvalid)
	}
	if err != nil {
		return AuthAccessResult{}, flow.technicalFailure(AuthEventLoginFailure, AuthReasonUserLookupFailed, err)
	}
	flow.setUser(user)

	err = s.comparePassword(ctx, input.Password, user.PasswordHash)
	if err != nil {
		return AuthAccessResult{}, flow.businessFailure(AuthEventLoginFailure, AuthReasonCredentialsInvalid, ErrCredentialsInvalid)
	}

	result, err := s.issueAccessResult(ctx, user)
	if err != nil {
		return AuthAccessResult{}, flow.technicalFailure(AuthEventTokenIssueFailure, AuthReasonAccessTokenIssueFail, err)
	}

	flow.success(AuthEventLoginSuccess)

	return result, nil
}

func (s *Service) RequestPasswordReset(ctx context.Context, input RequestPasswordResetInput) error {
	ctx, flow := s.startAuthFlow(ctx, "auth.password_reset.request", AuthActionResetRequest, input.Email)
	defer flow.end()

	input, err := normalizeRequestPasswordResetInput(input)
	if err != nil {
		return flow.businessFailure(AuthEventResetRequestFailure, AuthReasonInvalidInput, err)
	}

	if s.resets == nil || s.resetTokens == nil || s.resetMailer == nil {
		return flow.businessFailure(AuthEventResetRequestFailure, AuthReasonResetUnavailable, ErrResetUnavailable)
	}

	user, err := s.getUserByEmail(ctx, input.Email)
	if errors.Is(err, identity.ErrUserNotFound) {
		flow.success(AuthEventResetRequestSuccess)
		return nil
	}
	if err != nil {
		_ = flow.technicalFailure(AuthEventResetRequestFailure, AuthReasonUserLookupFailed, err)
		return nil
	}
	flow.setUser(user)

	rawToken, err := s.resetTokens.Generate(ctx)
	if err != nil {
		_ = flow.technicalFailure(AuthEventResetRequestFailure, AuthReasonResetTokenGenerateFail, err)
		return nil
	}

	now := s.now()
	expiresAt := now.Add(s.resetConfig.TokenTTL)
	if err := s.resets.InvalidateActivePasswordResetTokens(ctx, user.ID, now); err != nil {
		_ = flow.technicalFailure(AuthEventResetRequestFailure, AuthReasonResetTokenCreateFail, err)
		return nil
	}
	if _, err := s.resets.CreatePasswordResetToken(ctx, identity.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: s.resetTokens.Hash(rawToken),
		ExpiresAt: expiresAt,
	}); err != nil {
		_ = flow.technicalFailure(AuthEventResetRequestFailure, AuthReasonResetTokenCreateFail, err)
		return nil
	}

	if err := s.resetMailer.SendPasswordReset(ctx, identity.PasswordResetEmail{
		To:          user.Email,
		DisplayName: user.DisplayName,
		ResetURL:    passwordResetURL(input.ResetBaseURL, rawToken),
		ExpiresAt:   expiresAt,
	}); err != nil {
		_ = flow.technicalFailure(AuthEventResetRequestFailure, AuthReasonResetMailerFailed, err)
		return nil
	}

	flow.success(AuthEventResetRequestSuccess)

	return nil
}

func (s *Service) ConfirmPasswordReset(ctx context.Context, input ConfirmPasswordResetInput) error {
	ctx, flow := s.startAuthFlow(ctx, "auth.password_reset.confirm", AuthActionResetConfirm, "")
	defer flow.end()

	input, err := normalizeConfirmPasswordResetInput(input)
	if err != nil {
		return flow.businessFailure(AuthEventResetConfirmFailure, AuthReasonInvalidInput, err)
	}

	if s.resets == nil || s.resetTokens == nil {
		return flow.businessFailure(AuthEventResetConfirmFailure, AuthReasonResetUnavailable, ErrResetUnavailable)
	}

	passwordHash, err := s.hashPassword(ctx, input.NewPassword)
	if err != nil {
		return flow.technicalFailure(AuthEventResetConfirmFailure, AuthReasonPasswordHashFailed, err)
	}

	user, err := s.resets.ResetPasswordWithToken(ctx, s.resetTokens.Hash(input.Token), passwordHash, s.now())
	if errors.Is(err, identity.ErrResetTokenNotFound) {
		return flow.businessFailure(AuthEventResetConfirmFailure, AuthReasonResetTokenInvalid, ErrResetTokenInvalid)
	}
	if err != nil {
		return flow.technicalFailure(AuthEventResetConfirmFailure, AuthReasonPasswordResetFailed, err)
	}
	flow.setUser(user)
	flow.success(AuthEventResetConfirmSuccess)

	return nil
}

func passwordResetURL(baseURL, token string) string {
	values := url.Values{}
	values.Set("token", token)
	return baseURL + "/reset-password?" + values.Encode()
}

func (s *Service) issueAccessResult(ctx context.Context, user identity.User) (AuthAccessResult, error) {
	ctx, span := authTracer.Start(ctx, "auth.issue_access_token")
	defer span.End()

	token, err := s.tokens.IssueAccessToken(ctx, identity.AccessTokenClaims{
		Subject: user.ID,
		Email:   user.Email,
	})
	if err != nil {
		recordSpanError(span, err, AuthReasonAccessTokenIssueFail)
		return AuthAccessResult{}, err
	}

	return AuthAccessResult{
		UserID:      user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AccessToken: token.Value,
		ExpiresIn:   token.ExpiresIn,
	}, nil
}

func (s *Service) hashPassword(ctx context.Context, password string) (string, error) {
	_, span := authTracer.Start(ctx, "auth.password_hash")
	defer span.End()

	passwordHash, err := s.hasher.Hash(ctx, password)
	if err != nil {
		recordSpanError(span, err, AuthReasonPasswordHashFailed)
		return "", err
	}

	return passwordHash, nil
}

func (s *Service) comparePassword(ctx context.Context, password, passwordHash string) error {
	_, span := authTracer.Start(ctx, "auth.password_compare")
	defer span.End()

	return s.hasher.Compare(ctx, password, passwordHash)
}

func (s *Service) createUser(ctx context.Context, input identity.User) (identity.User, error) {
	ctx, span := authTracer.Start(ctx, "auth.create_user")
	defer span.End()

	user, err := s.users.CreateUser(ctx, input)
	if err != nil {
		if !errors.Is(err, identity.ErrEmailAlreadyExists) {
			recordSpanError(span, err, AuthReasonUserCreateFailed)
		}
		return identity.User{}, err
	}

	return user, nil
}

func (s *Service) getUserByEmail(ctx context.Context, email string) (identity.User, error) {
	ctx, span := authTracer.Start(ctx, "auth.get_user_by_email")
	defer span.End()

	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, identity.ErrUserNotFound) {
			recordSpanError(span, err, AuthReasonUserLookupFailed)
		}
		return identity.User{}, err
	}

	return user, nil
}

// GetCurrentUser fetches the live user record from the database using the
// stable user id embedded as the access token subject.
func (s *Service) GetCurrentUser(ctx context.Context, userID string) (identity.User, error) {
	ctx, span := authTracer.Start(ctx, "auth.get_current_user")
	defer span.End()

	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		recordSpanError(span, err, AuthReasonUserLookupFailed)
		return identity.User{}, err
	}

	return user, nil
}
