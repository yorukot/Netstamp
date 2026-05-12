package auth

import (
	"context"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

type UserRepository interface {
	CreateUser(ctx context.Context, input identity.User) (identity.User, error)
	GetUserByEmail(ctx context.Context, email string) (identity.User, error)
}

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Compare(ctx context.Context, password, passwordHash string) error
}

type TokenIssuer interface {
	IssueAccessToken(ctx context.Context, input identity.AccessTokenClaims) (identity.IssuedToken, error)
}

type TokenVerifier interface {
	VerifyAccessToken(ctx context.Context, value string) (identity.AccessTokenClaims, error)
}

// SecurityEventRecorder records security-relevant auth events.
type SecurityEventRecorder interface {
	RecordAuthEvent(ctx context.Context, event AuthEvent)
}

type AuthEventName string

const (
	AuthEventRegisterSuccess   AuthEventName = "auth.register.success"
	AuthEventRegisterFailure   AuthEventName = "auth.register.failure"
	AuthEventLoginSuccess      AuthEventName = "auth.login.success"
	AuthEventLoginFailure      AuthEventName = "auth.login.failure"
	AuthEventTokenIssueFailure AuthEventName = "auth.token.issue.failure" //nolint:gosec // Event names are not credentials.
)

type AuthEventAction string

const (
	AuthActionRegister AuthEventAction = "register"
	AuthActionLogin    AuthEventAction = "login"
)

type AuthEventOutcome string

const (
	AuthOutcomeSuccess AuthEventOutcome = "success"
	AuthOutcomeFailure AuthEventOutcome = "failure"
)

type AuthEventReason string

const (
	AuthReasonCredentialsInvalid   AuthEventReason = "credentials_invalid"
	AuthReasonEmailAlreadyExists   AuthEventReason = "email_already_exists"
	AuthReasonInvalidInput         AuthEventReason = "invalid_input"
	AuthReasonPasswordHashFailed   AuthEventReason = "password_hash_failed"
	AuthReasonUserCreateFailed     AuthEventReason = "user_create_failed"
	AuthReasonUserLookupFailed     AuthEventReason = "user_lookup_failed"
	AuthReasonAccessTokenIssueFail AuthEventReason = "access_token_issue_failed"
)

type AuthEvent struct {
	Name    AuthEventName
	Action  AuthEventAction
	Outcome AuthEventOutcome
	Reason  AuthEventReason
	UserID  string
	Email   string
	Err     error
}
