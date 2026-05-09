package auth

import (
	"context"
	"errors"
	"strings"
	"testing"

	appvalidation "github.com/yorukot/netstamp/internal/application/validation"
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
	repo := &fakeUserRepository{}
	hasher := &fakePasswordHasher{}
	service := NewService(
		repo,
		hasher,
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
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	assertValidationFieldError(t, err, "displayName", "must not be blank")

	assertRecordedEvent(t, recorder, AuthEvent{
		Name:    AuthEventRegisterFailure,
		Action:  AuthActionRegister,
		Outcome: AuthOutcomeFailure,
		Reason:  AuthReasonDisplayNameInvalid,
		Email:   "user@example.com",
	})
	if hasher.hashCalls != 0 {
		t.Fatalf("expected password not to be hashed for invalid input")
	}
	if repo.gotCreateInput.Email != "" {
		t.Fatalf("expected user not to be created for invalid input")
	}
}

func TestRegisterRecordsInvalidInputFailure(t *testing.T) {
	tests := []struct {
		name        string
		input       RegisterInput
		wantField   string
		wantMessage string
	}{
		{
			name: "blank email",
			input: RegisterInput{
				Email:       "   ",
				DisplayName: "Example User",
				Password:    "correct-password",
			},
			wantField:   "email",
			wantMessage: "must not be blank",
		},
		{
			name: "invalid email",
			input: RegisterInput{
				Email:       "not-an-email",
				DisplayName: "Example User",
				Password:    "correct-password",
			},
			wantField:   "email",
			wantMessage: "must be a valid email address",
		},
		{
			name: "email too long",
			input: RegisterInput{
				Email:       strings.Repeat("a", 245) + "@example.com",
				DisplayName: "Example User",
				Password:    "correct-password",
			},
			wantField:   "email",
			wantMessage: "must be at most 254 characters",
		},
		{
			name: "blank password",
			input: RegisterInput{
				Email:       "user@example.com",
				DisplayName: "Example User",
				Password:    "   ",
			},
			wantField:   "password",
			wantMessage: "must not be blank",
		},
		{
			name: "password too short",
			input: RegisterInput{
				Email:       "user@example.com",
				DisplayName: "Example User",
				Password:    "short",
			},
			wantField:   "password",
			wantMessage: "must be at least 8 characters",
		},
		{
			name: "password too long",
			input: RegisterInput{
				Email:       "user@example.com",
				DisplayName: "Example User",
				Password:    strings.Repeat("a", 129),
			},
			wantField:   "password",
			wantMessage: "must be at most 128 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := &recordingSecurityEventRecorder{}
			repo := &fakeUserRepository{}
			hasher := &fakePasswordHasher{}
			service := NewService(repo, hasher, &fakeTokenIssuer{}, recorder)

			_, err := service.Register(context.Background(), tt.input)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected invalid input, got %v", err)
			}
			assertValidationFieldError(t, err, tt.wantField, tt.wantMessage)
			assertRecordedEvent(t, recorder, AuthEvent{
				Name:    AuthEventRegisterFailure,
				Action:  AuthActionRegister,
				Outcome: AuthOutcomeFailure,
				Reason:  AuthReasonInvalidInput,
				Email:   strings.ToLower(strings.TrimSpace(tt.input.Email)),
			})
			if hasher.hashCalls != 0 {
				t.Fatalf("expected password not to be hashed for invalid input")
			}
			if repo.gotCreateInput.Email != "" {
				t.Fatalf("expected user not to be created for invalid input")
			}
		})
	}
}

func TestRegisterRecordsTooLongDisplayNameFailure(t *testing.T) {
	recorder := &recordingSecurityEventRecorder{}
	service := NewService(
		&fakeUserRepository{},
		&fakePasswordHasher{},
		&fakeTokenIssuer{},
		recorder,
	)

	_, err := service.Register(context.Background(), RegisterInput{
		Email:       "User@Example.COM",
		DisplayName: strings.Repeat("a", 101),
		Password:    "correct-password",
	})
	if !errors.Is(err, ErrDisplayNameTooLong) {
		t.Fatalf("expected display name too long, got %v", err)
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	assertValidationFieldError(t, err, "displayName", "must be at most 100 characters")

	assertRecordedEvent(t, recorder, AuthEvent{
		Name:    AuthEventRegisterFailure,
		Action:  AuthActionRegister,
		Outcome: AuthOutcomeFailure,
		Reason:  AuthReasonDisplayNameInvalid,
		Email:   "user@example.com",
	})
}

func TestRegisterNormalizesDisplayNameAndEmail(t *testing.T) {
	recorder := &recordingSecurityEventRecorder{}
	repo := &fakeUserRepository{}
	hasher := &fakePasswordHasher{}
	tokenIssuer := &fakeTokenIssuer{token: identity.IssuedToken{Value: "access-token", TokenType: "Bearer", ExpiresIn: 3600}}
	service := NewService(
		repo,
		hasher,
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
	if repo.gotCreateInput.Email != "user@example.com" {
		t.Fatalf("expected normalized email in create input, got %q", repo.gotCreateInput.Email)
	}
	if hasher.gotPassword != "correct-password" {
		t.Fatalf("expected password to be hashed unchanged, got %q", hasher.gotPassword)
	}
	if result.DisplayName == nil || *result.DisplayName != "Example User" {
		t.Fatalf("expected display name result, got %#v", result.DisplayName)
	}
	if tokenIssuer.gotInput.DisplayName == nil || *tokenIssuer.gotInput.DisplayName != "Example User" {
		t.Fatalf("expected display name in token input, got %#v", tokenIssuer.gotInput.DisplayName)
	}
}

func TestRegisterDoesNotTrimPasswordBeforeHashing(t *testing.T) {
	recorder := &recordingSecurityEventRecorder{}
	repo := &fakeUserRepository{}
	hasher := &fakePasswordHasher{}
	tokenIssuer := &fakeTokenIssuer{token: identity.IssuedToken{Value: "access-token", TokenType: "Bearer", ExpiresIn: 3600}}
	service := NewService(repo, hasher, tokenIssuer, recorder)

	_, err := service.Register(context.Background(), RegisterInput{
		Email:       "user@example.com",
		DisplayName: "Example User",
		Password:    "  correct-password  ",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if hasher.gotPassword != "  correct-password  " {
		t.Fatalf("expected original password value to be hashed, got %q", hasher.gotPassword)
	}
	if repo.gotCreateInput.PasswordHash != "hashed:  correct-password  " {
		t.Fatalf("expected hash of original password, got %q", repo.gotCreateInput.PasswordHash)
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

func assertValidationFieldError(t *testing.T, err error, wantField, wantMessage string) {
	t.Helper()

	fields, ok := appvalidation.FieldErrors(err)
	if !ok {
		t.Fatalf("expected validation field errors, got %v", err)
	}
	for _, fieldErr := range fields {
		if fieldErr.Field == wantField && fieldErr.Message == wantMessage {
			return
		}
	}

	t.Fatalf("expected field error %q/%q, got %#v", wantField, wantMessage, fields)
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
	hashErr     error
	compareErr  error
	hashCalls   int
	gotPassword string
}

func (h *fakePasswordHasher) Hash(password string) (string, error) {
	h.hashCalls++
	h.gotPassword = password
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
