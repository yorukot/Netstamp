package auth

import (
	"context"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

type UserRepository interface {
	CreateUser(ctx context.Context, input identity.User) (identity.User, error)
	GetUserByID(ctx context.Context, userID string) (identity.User, error)
	GetUserByEmail(ctx context.Context, email string) (identity.User, error)
	UpdateUserPasswordHash(ctx context.Context, input identity.User) (identity.User, error)
}

type SystemAdminRepository interface {
	GrantFirstSystemAdminIfNone(ctx context.Context, userID string) (bool, error)
}

type PasswordResetRepository interface {
	CreatePasswordResetToken(ctx context.Context, input identity.PasswordResetToken) (identity.PasswordResetToken, error)
	InvalidateActivePasswordResetTokens(ctx context.Context, userID string, usedAt time.Time) error
	GetActivePasswordResetTokenByHash(ctx context.Context, tokenHash string, now time.Time) (identity.PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, tokenID string, usedAt time.Time) error
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

type PasswordResetTokenManager interface {
	Generate(ctx context.Context) (string, error)
	Hash(value string) string
}

type PasswordResetMailer interface {
	SendPasswordReset(ctx context.Context, input identity.PasswordResetEmail) error
}

// SecurityEventRecorder records security-relevant auth events.
type SecurityEventRecorder interface {
	RecordAuthEvent(ctx context.Context, event AuthEvent)
}

type AuthEventName string

const (
	AuthEventRegisterSuccess     AuthEventName = "auth.register.success"
	AuthEventRegisterFailure     AuthEventName = "auth.register.failure"
	AuthEventLoginSuccess        AuthEventName = "auth.login.success"
	AuthEventLoginFailure        AuthEventName = "auth.login.failure"
	AuthEventTokenIssueFailure   AuthEventName = "auth.token.issue.failure" //nolint:gosec // Event names are not credentials.
	AuthEventResetRequestSuccess AuthEventName = "auth.password_reset.request.success"
	AuthEventResetRequestFailure AuthEventName = "auth.password_reset.request.failure"
	AuthEventResetConfirmSuccess AuthEventName = "auth.password_reset.confirm.success"
	AuthEventResetConfirmFailure AuthEventName = "auth.password_reset.confirm.failure"
)

type AuthEventAction string

const (
	AuthActionRegister     AuthEventAction = "register"
	AuthActionLogin        AuthEventAction = "login"
	AuthActionResetRequest AuthEventAction = "password_reset.request"
	AuthActionResetConfirm AuthEventAction = "password_reset.confirm"
)

type AuthEventOutcome string

const (
	AuthOutcomeSuccess AuthEventOutcome = "success"
	AuthOutcomeFailure AuthEventOutcome = "failure"
)

type AuthEventReason string

const (
	AuthReasonCredentialsInvalid     AuthEventReason = "credentials_invalid"
	AuthReasonEmailAlreadyExists     AuthEventReason = "email_already_exists"
	AuthReasonInvalidInput           AuthEventReason = "invalid_input"
	AuthReasonPasswordHashFailed     AuthEventReason = "password_hash_failed"
	AuthReasonResetMailerFailed      AuthEventReason = "password_reset_mail_failed"
	AuthReasonResetTokenCreateFail   AuthEventReason = "password_reset_token_create_failed"
	AuthReasonResetTokenGenerateFail AuthEventReason = "password_reset_token_generate_failed"
	AuthReasonResetTokenInvalid      AuthEventReason = "password_reset_token_invalid"
	AuthReasonResetUnavailable       AuthEventReason = "password_reset_unavailable"
	AuthReasonPasswordResetFailed    AuthEventReason = "password_reset_failed"
	AuthReasonUserCreateFailed       AuthEventReason = "user_create_failed"
	AuthReasonUserLookupFailed       AuthEventReason = "user_lookup_failed"
	AuthReasonAccessTokenIssueFail   AuthEventReason = "access_token_issue_failed"
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
