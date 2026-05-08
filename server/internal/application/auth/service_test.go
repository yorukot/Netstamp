package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestLoginRecordsSuccess(t *testing.T) {
	recorder := &recordingSecurityEventRecorder{}
	tokenIssuer := &fakeTokenIssuer{token: identity.IssuedToken{Value: "access-token", TokenType: "Bearer", ExpiresIn: 3600}}
	repo := &fakeUserRepository{
		user: identity.User{
			ID:           "user-1",
			Email:        "user@example.com",
			DisplayName:  stringPtr("Example User"),
			PasswordHash: "password-hash",
			IsActive:     true,
		},
	}
	service := NewService(
		repo,
		&fakePasswordHasher{},
		tokenIssuer,
		recorder,
	)

	result, err := service.Login(context.Background(), LoginInput{
		Email:    " User@Example.COM ",
		Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if repo.gotEmail != "user@example.com" {
		t.Fatalf("expected normalized lookup email, got %q", repo.gotEmail)
	}
	if result.UserID != "user-1" {
		t.Fatalf("expected user id, got %q", result.UserID)
	}
	if result.DisplayName == nil || *result.DisplayName != "Example User" {
		t.Fatalf("expected display name, got %#v", result.DisplayName)
	}
	if tokenIssuer.gotInput.DisplayName == nil || *tokenIssuer.gotInput.DisplayName != "Example User" {
		t.Fatalf("expected display name in token input, got %#v", tokenIssuer.gotInput.DisplayName)
	}
	assertRecordedEvent(t, recorder, AuthEvent{
		Name:    AuthEventLoginSuccess,
		Action:  AuthActionLogin,
		Outcome: AuthOutcomeSuccess,
		UserID:  "user-1",
		Email:   "user@example.com",
	})
}

func TestLoginRecordsInvalidCredentialFailure(t *testing.T) {
	recorder := &recordingSecurityEventRecorder{}
	service := NewService(
		&fakeUserRepository{getErr: identity.ErrUserNotFound},
		&fakePasswordHasher{},
		&fakeTokenIssuer{},
		recorder,
	)

	_, err := service.Login(context.Background(), LoginInput{
		Email:    " Missing@Example.COM ",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrCredentialsInvalid) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}

	assertRecordedEvent(t, recorder, AuthEvent{
		Name:    AuthEventLoginFailure,
		Action:  AuthActionLogin,
		Outcome: AuthOutcomeFailure,
		Reason:  AuthReasonCredentialsInvalid,
		Email:   "missing@example.com",
	})
}

func TestLoginRecordsInactiveUserFailure(t *testing.T) {
	recorder := &recordingSecurityEventRecorder{}
	service := NewService(
		&fakeUserRepository{
			user: identity.User{
				ID:           "user-1",
				Email:        "user@example.com",
				PasswordHash: "password-hash",
				IsActive:     false,
			},
		},
		&fakePasswordHasher{},
		&fakeTokenIssuer{},
		recorder,
	)

	_, err := service.Login(context.Background(), LoginInput{
		Email:    "User@Example.COM",
		Password: "correct-password",
	})
	if !errors.Is(err, ErrUserInactive) {
		t.Fatalf("expected inactive user, got %v", err)
	}

	assertRecordedEvent(t, recorder, AuthEvent{
		Name:    AuthEventLoginFailure,
		Action:  AuthActionLogin,
		Outcome: AuthOutcomeFailure,
		Reason:  AuthReasonUserInactive,
		UserID:  "user-1",
		Email:   "user@example.com",
	})
}

func TestRegisterRecordsDuplicateEmailFailure(t *testing.T) {
	recorder := &recordingSecurityEventRecorder{}
	service := NewService(
		&fakeUserRepository{createErr: ErrEmailAlreadyExists},
		&fakePasswordHasher{},
		&fakeTokenIssuer{},
		recorder,
	)

	_, err := service.Register(context.Background(), RegisterInput{
		Email:       "Existing@Example.COM",
		DisplayName: "Existing User",
		Password:    "correct-password",
	})
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("expected duplicate email, got %v", err)
	}

	assertRecordedEvent(t, recorder, AuthEvent{
		Name:    AuthEventRegisterFailure,
		Action:  AuthActionRegister,
		Outcome: AuthOutcomeFailure,
		Reason:  AuthReasonEmailAlreadyExists,
		Email:   "existing@example.com",
	})
}

func TestRegisterRecordsInvalidDisplayNameFailure(t *testing.T) {
	recorder := &recordingSecurityEventRecorder{}
	service := NewService(
		&fakeUserRepository{},
		&fakePasswordHasher{},
		&fakeTokenIssuer{},
		recorder,
	)

	_, err := service.Register(context.Background(), RegisterInput{
		Email:       "User@Example.COM",
		DisplayName: "   ",
		Password:    "correct-password",
	})
	if !errors.Is(err, ErrDisplayNameRequired) {
		t.Fatalf("expected display name required, got %v", err)
	}

	assertRecordedEvent(t, recorder, AuthEvent{
		Name:    AuthEventRegisterFailure,
		Action:  AuthActionRegister,
		Outcome: AuthOutcomeFailure,
		Reason:  AuthReasonDisplayNameInvalid,
		Email:   "user@example.com",
	})
}

func TestRegisterNormalizesDisplayName(t *testing.T) {
	recorder := &recordingSecurityEventRecorder{}
	repo := &fakeUserRepository{}
	tokenIssuer := &fakeTokenIssuer{token: identity.IssuedToken{Value: "access-token", TokenType: "Bearer", ExpiresIn: 3600}}
	service := NewService(
		repo,
		&fakePasswordHasher{},
		tokenIssuer,
		recorder,
	)

	result, err := service.Register(context.Background(), RegisterInput{
		Email:       "User@Example.COM",
		DisplayName: "  Example User  ",
		Password:    "correct-password",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if repo.gotCreateInput.DisplayName != "Example User" {
		t.Fatalf("expected normalized display name in create input, got %q", repo.gotCreateInput.DisplayName)
	}
	if result.DisplayName == nil || *result.DisplayName != "Example User" {
		t.Fatalf("expected display name result, got %#v", result.DisplayName)
	}
	if tokenIssuer.gotInput.DisplayName == nil || *tokenIssuer.gotInput.DisplayName != "Example User" {
		t.Fatalf("expected display name in token input, got %#v", tokenIssuer.gotInput.DisplayName)
	}
}

func TestRegisterRecordsTokenIssueFailure(t *testing.T) {
	recorder := &recordingSecurityEventRecorder{}
	tokenErr := errors.New("sign token")
	service := NewService(
		&fakeUserRepository{
			createdUser: identity.User{
				ID:          "user-1",
				Email:       "user@example.com",
				DisplayName: stringPtr("Example User"),
				IsActive:    true,
			},
		},
		&fakePasswordHasher{},
		&fakeTokenIssuer{err: tokenErr},
		recorder,
	)

	_, err := service.Register(context.Background(), RegisterInput{
		Email:       "User@Example.COM",
		DisplayName: "Example User",
		Password:    "correct-password",
	})
	if !errors.Is(err, tokenErr) {
		t.Fatalf("expected token error, got %v", err)
	}

	assertRecordedEvent(t, recorder, AuthEvent{
		Name:    AuthEventTokenIssueFailure,
		Action:  AuthActionRegister,
		Outcome: AuthOutcomeFailure,
		Reason:  AuthReasonAccessTokenIssueFail,
		UserID:  "user-1",
		Email:   "user@example.com",
		Err:     tokenErr,
	})
}

func assertRecordedEvent(t *testing.T, recorder *recordingSecurityEventRecorder, want AuthEvent) {
	t.Helper()

	if len(recorder.events) != 1 {
		t.Fatalf("expected one event, got %d: %#v", len(recorder.events), recorder.events)
	}

	got := recorder.events[0]
	if got.Name != want.Name ||
		got.Action != want.Action ||
		got.Outcome != want.Outcome ||
		got.Reason != want.Reason ||
		got.UserID != want.UserID ||
		got.Email != want.Email ||
		!errors.Is(got.Err, want.Err) {
		t.Fatalf("unexpected event:\n got: %#v\nwant: %#v", got, want)
	}
}

type recordingSecurityEventRecorder struct {
	events []AuthEvent
}

func (r *recordingSecurityEventRecorder) RecordAuthEvent(_ context.Context, event AuthEvent) {
	r.events = append(r.events, event)
}

type fakeUserRepository struct {
	user           identity.User
	createdUser    identity.User
	getErr         error
	createErr      error
	gotEmail       string
	gotCreateInput identity.CreateUserInput
}

func (r *fakeUserRepository) CreateUser(_ context.Context, input identity.CreateUserInput) (identity.User, error) {
	r.gotCreateInput = input
	if r.createErr != nil {
		return identity.User{}, r.createErr
	}
	if r.createdUser.ID != "" {
		return r.createdUser, nil
	}
	return identity.User{
		ID:           "created-user",
		Email:        input.Email,
		DisplayName:  &input.DisplayName,
		PasswordHash: input.PasswordHash,
		IsActive:     true,
	}, nil
}

func (r *fakeUserRepository) GetUserByEmail(_ context.Context, email string) (identity.User, error) {
	r.gotEmail = email
	if r.getErr != nil {
		return identity.User{}, r.getErr
	}
	return r.user, nil
}

type fakePasswordHasher struct {
	hashErr    error
	compareErr error
}

func (h *fakePasswordHasher) Hash(password string) (string, error) {
	if h.hashErr != nil {
		return "", h.hashErr
	}
	return "hashed:" + password, nil
}

func (h *fakePasswordHasher) Compare(_, _ string) error {
	return h.compareErr
}

type fakeTokenIssuer struct {
	token    identity.IssuedToken
	err      error
	gotInput identity.AccessTokenInput
}

func (i *fakeTokenIssuer) IssueAccessToken(_ context.Context, input identity.AccessTokenInput) (identity.IssuedToken, error) {
	i.gotInput = input
	if i.err != nil {
		return identity.IssuedToken{}, i.err
	}
	return i.token, nil
}

func stringPtr(value string) *string {
	return &value
}
