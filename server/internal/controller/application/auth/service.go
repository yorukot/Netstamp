package auth

import (
	"context"
	"errors"
	"net/url"
	"time"

	apptx "github.com/yorukot/netstamp/internal/controller/application/tx"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

type Service struct {
	users                   UserRepository
	systemAdmin             SystemAdminRepository
	hasher                  PasswordHasher
	sessions                SessionManager
	events                  SecurityEventRecorder
	resets                  PasswordResetRepository
	resetTokens             PasswordResetTokenManager
	resetMailer             PasswordResetMailer
	resetConfig             PasswordResetConfig
	emailVerifications      EmailVerificationRepository
	emailVerificationTokens EmailVerificationTokenManager
	emailVerificationMailer EmailVerificationMailer
	emailVerificationConfig EmailVerificationConfig
	tx                      apptx.Transactor
	now                     func() time.Time
}

func NewService(users UserRepository, hasher PasswordHasher, sessions SessionManager, events SecurityEventRecorder, transactors ...apptx.Transactor) *Service {
	tx := apptx.Transactor(apptx.NoopTransactor{})
	if len(transactors) > 0 && transactors[0] != nil {
		tx = transactors[0]
	}

	return &Service{
		users:                   users,
		hasher:                  hasher,
		sessions:                sessions,
		events:                  events,
		resetConfig:             PasswordResetConfig{TokenTTL: 30 * time.Minute},
		emailVerificationConfig: EmailVerificationConfig{TokenTTL: 24 * time.Hour},
		tx:                      tx,
		now:                     func() time.Time { return time.Now().UTC() },
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

func (s *Service) ConfigureSystemAdmin(repo SystemAdminRepository) {
	s.systemAdmin = repo
}

func (s *Service) ConfigureEmailVerification(verifications EmailVerificationRepository, tokens EmailVerificationTokenManager, mailer EmailVerificationMailer, cfg EmailVerificationConfig) {
	s.emailVerifications = verifications
	s.emailVerificationTokens = tokens
	s.emailVerificationMailer = mailer
	if cfg.TokenTTL <= 0 {
		cfg.TokenTTL = 24 * time.Hour
	}
	s.emailVerificationConfig = cfg
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

	now := s.now()
	created, err := s.createRegisteredUser(ctx, input, passwordHash, now)
	if errors.Is(err, identity.ErrEmailAlreadyExists) {
		return AuthAccessResult{}, flow.businessFailure(AuthEventRegisterFailure, AuthReasonEmailAlreadyExists, err)
	}
	if err != nil {
		return AuthAccessResult{}, flow.technicalFailure(AuthEventRegisterFailure, AuthReasonUserCreateFailed, err)
	}
	user := created.user
	flow.setUser(user)

	if input.RequireEmailVerification && user.EmailVerifiedAt == nil {
		sendErr := s.sendEmailVerification(ctx, user, input.EmailVerificationBaseURL, created.rawEmailVerificationToken, now.Add(s.emailVerificationConfig.TokenTTL))
		if sendErr != nil {
			flow.recordTechnicalFailure(AuthEventRegisterFailure, AuthReasonEmailVerificationMailerFailed, sendErr)
		}
		flow.success(AuthEventRegisterSuccess)
		return AuthAccessResult{
			UserID:                    user.ID,
			Email:                     user.Email,
			DisplayName:               user.DisplayName,
			EmailVerified:             false,
			IsSystemAdmin:             user.IsSystemAdmin,
			EmailVerificationRequired: true,
		}, nil
	}

	result, err := s.createAccessResult(ctx, user)
	if err != nil {
		return AuthAccessResult{}, flow.technicalFailure(AuthEventSessionCreateFailure, AuthReasonSessionCreateFail, err)
	}

	flow.success(AuthEventRegisterSuccess)

	return result, nil
}

type registeredUser struct {
	user                      identity.User
	rawEmailVerificationToken string
}

func (s *Service) createRegisteredUser(ctx context.Context, input RegisterInput, passwordHash string, now time.Time) (registeredUser, error) {
	var result registeredUser
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		created, err := s.createUser(ctx, identity.User{
			Email:           input.Email,
			DisplayName:     input.DisplayName,
			PasswordHash:    passwordHash,
			EmailVerifiedAt: initialEmailVerifiedAt(input.RequireEmailVerification, now),
		})
		if err != nil {
			return err
		}

		created, err = s.applyFirstSystemAdminGrant(ctx, created, now)
		if err != nil {
			return err
		}

		rawToken, err := s.maybeCreateRegistrationEmailVerificationToken(ctx, input, created, now)
		if err != nil {
			return err
		}
		result = registeredUser{
			user:                      created,
			rawEmailVerificationToken: rawToken,
		}
		return nil
	})
	return result, err
}

func initialEmailVerifiedAt(requireEmailVerification bool, now time.Time) *time.Time {
	if requireEmailVerification {
		return nil
	}
	return &now
}

func (s *Service) applyFirstSystemAdminGrant(ctx context.Context, user identity.User, now time.Time) (identity.User, error) {
	if s.systemAdmin == nil {
		return user, nil
	}
	granted, err := s.systemAdmin.GrantFirstSystemAdminIfNone(ctx, user.ID)
	if err != nil {
		return identity.User{}, err
	}
	user.IsSystemAdmin = granted
	if !granted || user.EmailVerifiedAt != nil {
		return user, nil
	}

	verified, err := s.markUserEmailVerified(ctx, user.ID, now)
	if err != nil {
		return identity.User{}, err
	}
	verified.IsSystemAdmin = true
	return verified, nil
}

func (s *Service) maybeCreateRegistrationEmailVerificationToken(ctx context.Context, input RegisterInput, user identity.User, now time.Time) (string, error) {
	if !input.RequireEmailVerification || user.EmailVerifiedAt != nil {
		return "", nil
	}
	return s.createEmailVerificationToken(ctx, user.ID, now)
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

	if user.DisabledAt != nil {
		return AuthAccessResult{}, flow.businessFailure(AuthEventLoginFailure, AuthReasonAccountDisabled, ErrCredentialsInvalid)
	}

	err = s.comparePassword(ctx, input.Password, user.PasswordHash)
	if err != nil {
		return AuthAccessResult{}, flow.businessFailure(AuthEventLoginFailure, AuthReasonCredentialsInvalid, ErrCredentialsInvalid)
	}
	if input.RequireEmailVerification && user.EmailVerifiedAt == nil {
		return AuthAccessResult{}, flow.businessFailure(AuthEventLoginFailure, AuthReasonEmailVerificationRequired, ErrEmailVerificationRequired)
	}

	result, err := s.createAccessResult(ctx, user)
	if err != nil {
		return AuthAccessResult{}, flow.technicalFailure(AuthEventSessionCreateFailure, AuthReasonSessionCreateFail, err)
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
		flow.recordTechnicalFailure(AuthEventResetRequestFailure, AuthReasonUserLookupFailed, err)
		return nil
	}
	flow.setUser(user)
	if user.DisabledAt != nil {
		flow.success(AuthEventResetRequestSuccess)
		return nil
	}

	rawToken, err := s.resetTokens.Generate(ctx)
	if err != nil {
		flow.recordTechnicalFailure(AuthEventResetRequestFailure, AuthReasonResetTokenGenerateFail, err)
		return nil
	}

	now := s.now()
	expiresAt := now.Add(s.resetConfig.TokenTTL)
	if err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.resets.InvalidateActivePasswordResetTokens(ctx, user.ID, now); err != nil {
			return err
		}
		_, err := s.resets.CreatePasswordResetToken(ctx, identity.PasswordResetToken{
			UserID:    user.ID,
			TokenHash: s.resetTokens.Hash(rawToken),
			ExpiresAt: expiresAt,
		})
		return err
	}); err != nil {
		flow.recordTechnicalFailure(AuthEventResetRequestFailure, AuthReasonResetTokenCreateFail, err)
		return nil
	}

	if err := s.resetMailer.SendPasswordReset(ctx, identity.PasswordResetEmail{
		To:          user.Email,
		DisplayName: user.DisplayName,
		ResetURL:    passwordResetURL(input.ResetBaseURL, rawToken),
		ExpiresAt:   expiresAt,
	}); err != nil {
		flow.recordTechnicalFailure(AuthEventResetRequestFailure, AuthReasonResetMailerFailed, err)
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

	now := s.now()
	var user identity.User
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		token, tokenErr := s.resets.GetActivePasswordResetTokenByHash(ctx, s.resetTokens.Hash(input.Token), now)
		if tokenErr != nil {
			return tokenErr
		}
		updated, updateErr := s.users.UpdateUserPasswordHash(ctx, identity.User{
			ID:           token.UserID,
			PasswordHash: passwordHash,
		})
		if updateErr != nil {
			return updateErr
		}
		if markErr := s.resets.MarkPasswordResetTokenUsed(ctx, token.ID, now); markErr != nil {
			return markErr
		}
		if s.sessions != nil {
			if revokeErr := s.sessions.RevokeUserSessions(ctx, token.UserID, "password_reset"); revokeErr != nil {
				return revokeErr
			}
		}
		user = updated
		return nil
	})
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

func (s *Service) RequestEmailVerification(ctx context.Context, input RequestEmailVerificationInput) error {
	ctx, flow := s.startAuthFlow(ctx, "auth.email_verification.request", AuthActionEmailVerificationRequest, input.Email)
	defer flow.end()

	input, err := normalizeRequestEmailVerificationInput(input)
	if err != nil {
		return flow.businessFailure(AuthEventEmailVerificationRequestFailure, AuthReasonInvalidInput, err)
	}
	if s.emailVerifications == nil || s.emailVerificationTokens == nil || s.emailVerificationMailer == nil {
		return flow.businessFailure(AuthEventEmailVerificationRequestFailure, AuthReasonEmailVerificationUnavailable, ErrEmailVerificationUnavailable)
	}

	user, err := s.getUserByEmail(ctx, input.Email)
	if errors.Is(err, identity.ErrUserNotFound) {
		flow.success(AuthEventEmailVerificationRequestSuccess)
		return nil
	}
	if err != nil {
		flow.recordTechnicalFailure(AuthEventEmailVerificationRequestFailure, AuthReasonUserLookupFailed, err)
		return nil
	}
	flow.setUser(user)
	if user.EmailVerifiedAt != nil {
		flow.success(AuthEventEmailVerificationRequestSuccess)
		return nil
	}

	now := s.now()
	rawToken, err := s.createEmailVerificationToken(ctx, user.ID, now)
	if err != nil {
		flow.recordTechnicalFailure(AuthEventEmailVerificationRequestFailure, AuthReasonEmailVerificationTokenCreateFail, err)
		return nil
	}
	if err := s.sendEmailVerification(ctx, user, input.EmailVerificationBaseURL, rawToken, now.Add(s.emailVerificationConfig.TokenTTL)); err != nil {
		flow.recordTechnicalFailure(AuthEventEmailVerificationRequestFailure, AuthReasonEmailVerificationMailerFailed, err)
		return nil
	}

	flow.success(AuthEventEmailVerificationRequestSuccess)
	return nil
}

func (s *Service) ConfirmEmailVerification(ctx context.Context, input ConfirmEmailVerificationInput) error {
	ctx, flow := s.startAuthFlow(ctx, "auth.email_verification.confirm", AuthActionEmailVerificationConfirm, "")
	defer flow.end()

	input, err := normalizeConfirmEmailVerificationInput(input)
	if err != nil {
		return flow.businessFailure(AuthEventEmailVerificationConfirmFailure, AuthReasonInvalidInput, err)
	}
	if s.emailVerifications == nil || s.emailVerificationTokens == nil {
		return flow.businessFailure(AuthEventEmailVerificationConfirmFailure, AuthReasonEmailVerificationUnavailable, ErrEmailVerificationUnavailable)
	}

	now := s.now()
	var user identity.User
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		token, tokenErr := s.emailVerifications.GetActiveEmailVerificationTokenByHash(ctx, s.emailVerificationTokens.Hash(input.Token), now)
		if tokenErr != nil {
			return tokenErr
		}
		verified, verifyErr := s.emailVerifications.MarkUserEmailVerified(ctx, token.UserID, now)
		if verifyErr != nil {
			return verifyErr
		}
		if markErr := s.emailVerifications.MarkEmailVerificationTokenUsed(ctx, token.ID, now); markErr != nil {
			return markErr
		}
		user = verified
		return nil
	})
	if errors.Is(err, identity.ErrEmailVerificationTokenNotFound) {
		return flow.businessFailure(AuthEventEmailVerificationConfirmFailure, AuthReasonEmailVerificationTokenInvalid, ErrEmailVerificationTokenInvalid)
	}
	if err != nil {
		return flow.technicalFailure(AuthEventEmailVerificationConfirmFailure, AuthReasonEmailVerificationFailed, err)
	}
	flow.setUser(user)
	flow.success(AuthEventEmailVerificationConfirmSuccess)
	return nil
}

func passwordResetURL(baseURL, token string) string {
	values := url.Values{}
	values.Set("token", token)
	return baseURL + "/reset-password?" + values.Encode()
}

func emailVerificationURL(baseURL, token string) string {
	values := url.Values{}
	values.Set("token", token)
	return baseURL + "/verify-email?" + values.Encode()
}

func (s *Service) createAccessResult(ctx context.Context, user identity.User) (AuthAccessResult, error) {
	ctx, span := authTracer.Start(ctx, "auth.create_session")
	defer span.End()

	if s.sessions == nil {
		return AuthAccessResult{}, ErrSessionInvalid
	}

	session, err := s.sessions.CreateSession(ctx, CreateSessionInput{
		UserID: user.ID,
		Now:    s.now(),
	})
	if err != nil {
		recordSpanError(span, err, AuthReasonSessionCreateFail)
		return AuthAccessResult{}, err
	}

	return AuthAccessResult{
		UserID:        user.ID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		EmailVerified: user.EmailVerifiedAt != nil,
		IsSystemAdmin: user.IsSystemAdmin,
		SessionToken:  session.RawToken,
		ExpiresIn:     session.ExpiresIn,
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

func (s *Service) createEmailVerificationToken(ctx context.Context, userID string, now time.Time) (string, error) {
	if s.emailVerifications == nil || s.emailVerificationTokens == nil {
		return "", ErrEmailVerificationUnavailable
	}
	rawToken, err := s.emailVerificationTokens.Generate(ctx)
	if err != nil {
		return "", err
	}
	invalidateErr := s.emailVerifications.InvalidateActiveEmailVerificationTokens(ctx, userID, now)
	if invalidateErr != nil {
		return "", invalidateErr
	}
	_, err = s.emailVerifications.CreateEmailVerificationToken(ctx, identity.EmailVerificationToken{
		UserID:    userID,
		TokenHash: s.emailVerificationTokens.Hash(rawToken),
		ExpiresAt: now.Add(s.emailVerificationConfig.TokenTTL),
	})
	if err != nil {
		return "", err
	}
	return rawToken, nil
}

func (s *Service) sendEmailVerification(ctx context.Context, user identity.User, baseURL, rawToken string, expiresAt time.Time) error {
	if s.emailVerificationMailer == nil {
		return ErrEmailVerificationUnavailable
	}
	if baseURL == "" || rawToken == "" {
		return ErrEmailVerificationUnavailable
	}
	return s.emailVerificationMailer.SendEmailVerification(ctx, identity.EmailVerificationEmail{
		To:              user.Email,
		DisplayName:     user.DisplayName,
		VerificationURL: emailVerificationURL(baseURL, rawToken),
		ExpiresAt:       expiresAt,
	})
}

func (s *Service) markUserEmailVerified(ctx context.Context, userID string, verifiedAt time.Time) (identity.User, error) {
	if s.emailVerifications == nil {
		return identity.User{}, ErrEmailVerificationUnavailable
	}
	return s.emailVerifications.MarkUserEmailVerified(ctx, userID, verifiedAt)
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
// stable user id associated with the verified server-side session.
func (s *Service) GetCurrentUser(ctx context.Context, userID string) (identity.User, error) {
	ctx, span := authTracer.Start(ctx, "auth.get_current_user")
	defer span.End()

	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		recordSpanError(span, err, AuthReasonUserLookupFailed)
		return identity.User{}, err
	}
	if user.DisabledAt != nil {
		return identity.User{}, ErrAccountDisabled
	}

	return user, nil
}
