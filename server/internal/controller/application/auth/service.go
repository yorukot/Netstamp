package auth

import (
	"context"
	"errors"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

type Service struct {
	users  UserRepository
	hasher PasswordHasher
	tokens TokenIssuer
	events SecurityEventRecorder
}

func NewService(users UserRepository, hasher PasswordHasher, tokens TokenIssuer, events SecurityEventRecorder) *Service {
	return &Service{
		users:  users,
		hasher: hasher,
		tokens: tokens,
		events: events,
	}
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
	if errors.Is(err, ErrEmailAlreadyExists) {
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
	if errors.Is(err, ErrUserNotFound) {
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

func (s *Service) issueAccessResult(ctx context.Context, user identity.User) (AuthAccessResult, error) {
	ctx, span := authTracer.Start(ctx, "auth.issue_access_token")
	defer span.End()

	token, err := s.tokens.IssueAccessToken(ctx, identity.AccessTokenClaims{
		Subject:     user.ID,
		Email:       user.Email,
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
		TokenType:   token.TokenType,
		ExpiresIn:   token.ExpiresIn,
	}, nil
}

func (s *Service) hashPassword(ctx context.Context, password string) (string, error) {
	_, span := authTracer.Start(ctx, "auth.password_hash")
	defer span.End()

	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		recordSpanError(span, err, AuthReasonPasswordHashFailed)
		return "", err
	}

	return passwordHash, nil
}

func (s *Service) comparePassword(ctx context.Context, password, passwordHash string) error {
	_, span := authTracer.Start(ctx, "auth.password_compare")
	defer span.End()

	return s.hasher.Compare(password, passwordHash)
}

func (s *Service) createUser(ctx context.Context, input identity.User) (identity.User, error) {
	ctx, span := authTracer.Start(ctx, "auth.create_user")
	defer span.End()

	user, err := s.users.CreateUser(ctx, input)
	if err != nil {
		if !errors.Is(err, ErrEmailAlreadyExists) {
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
		if !errors.Is(err, ErrUserNotFound) {
			recordSpanError(span, err, AuthReasonUserLookupFailed)
		}
		return identity.User{}, err
	}

	return user, nil
}
