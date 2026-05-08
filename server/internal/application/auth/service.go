package auth

import (
	"context"
	"errors"
	"unicode/utf8"

	"github.com/yorukot/netstamp/internal/domain/identity"
	"github.com/yorukot/netstamp/internal/normalize"
)

const maxDisplayNameLength = 100

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
	email := normalize.Email(input.Email)
	displayName, displayNameErr := normalize.RequiredString(input.DisplayName, ErrDisplayNameRequired)
	ctx, flow := s.startAuthFlow(ctx, "auth.register", AuthActionRegister, email)
	defer flow.end()

	if displayNameErr != nil {
		return AuthAccessResult{}, flow.businessFailure(AuthEventRegisterFailure, AuthReasonDisplayNameInvalid, displayNameErr)
	}
	if utf8.RuneCountInString(displayName) > maxDisplayNameLength {
		return AuthAccessResult{}, flow.businessFailure(AuthEventRegisterFailure, AuthReasonDisplayNameInvalid, ErrDisplayNameTooLong)
	}

	passwordHash, err := s.hashPassword(ctx, input.Password)
	if err != nil {
		return AuthAccessResult{}, flow.technicalFailure(AuthEventRegisterFailure, AuthReasonPasswordHashFailed, err)
	}

	user, err := s.createUser(ctx, identity.CreateUserInput{
		Email:        email,
		DisplayName:  displayName,
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
	email := normalize.Email(input.Email)
	ctx, flow := s.startAuthFlow(ctx, "auth.login", AuthActionLogin, email)
	defer flow.end()

	user, err := s.getUserByEmail(ctx, email)
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

	if !user.IsActive {
		return AuthAccessResult{}, flow.businessFailure(AuthEventLoginFailure, AuthReasonUserInactive, ErrUserInactive)
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

	token, err := s.tokens.IssueAccessToken(ctx, identity.AccessTokenInput{
		Subject:     user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
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

func (s *Service) createUser(ctx context.Context, input identity.CreateUserInput) (identity.User, error) {
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
		if !errors.Is(err, identity.ErrUserNotFound) {
			recordSpanError(span, err, AuthReasonUserLookupFailed)
		}
		return identity.User{}, err
	}

	return user, nil
}
