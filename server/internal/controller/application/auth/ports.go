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

type OIDCRepository interface {
	CreateOIDCUser(ctx context.Context, email, displayName, issuer, subject string, now time.Time) (identity.User, identity.UserIdentity, error)
	CreateUserIdentity(ctx context.Context, input identity.UserIdentity) (identity.UserIdentity, error)
	GetUserIdentityByIssuerSubject(ctx context.Context, issuer, subject string) (identity.UserIdentity, error)
	GetUserIdentityByIDForUser(ctx context.Context, identityID, userID string) (identity.UserIdentity, error)
	ListUserIdentities(ctx context.Context, userID string) ([]identity.UserIdentity, error)
	TouchUserIdentityLogin(ctx context.Context, input identity.UserIdentity, at time.Time) (identity.UserIdentity, error)
	CreateOIDCAuthFlow(ctx context.Context, input identity.OIDCAuthFlow) (identity.OIDCAuthFlow, error)
	ConsumeOIDCAuthFlow(ctx context.Context, stateHash, browserTokenHash []byte, now time.Time) (identity.OIDCAuthFlow, error)
	DeleteExpiredOIDCAuthFlows(ctx context.Context, now time.Time) error
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

type EmailVerificationRepository interface {
	CreateEmailVerificationToken(ctx context.Context, input identity.EmailVerificationToken) (identity.EmailVerificationToken, error)
	InvalidateActiveEmailVerificationTokens(ctx context.Context, userID string, usedAt time.Time) error
	GetActiveEmailVerificationTokenByHash(ctx context.Context, tokenHash string, now time.Time) (identity.EmailVerificationToken, error)
	MarkEmailVerificationTokenUsed(ctx context.Context, tokenID string, usedAt time.Time) error
	MarkUserEmailVerified(ctx context.Context, userID string, verifiedAt time.Time) (identity.User, error)
}

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Compare(ctx context.Context, password, passwordHash string) error
}

type SessionManager interface {
	CreateSession(ctx context.Context, input CreateSessionInput) (identity.CreatedSession, error)
	VerifySession(ctx context.Context, rawToken string) (identity.SessionClaims, error)
	CreateCSRFToken(ctx context.Context, sessionID string) (string, error)
	VerifyCSRFToken(ctx context.Context, sessionID, rawToken string) error
	RevokeSession(ctx context.Context, rawToken, reason string) error
	ListUserSessions(ctx context.Context, userID string) ([]identity.AuthSession, error)
	RevokeUserSession(ctx context.Context, userID, sessionID, reason string) error
	RevokeUserSessions(ctx context.Context, userID, reason string) error
}

type RecentAuthenticationManager interface {
	SudoStatus(ctx context.Context, sessionID string) (identity.SudoStatus, error)
	ElevateSession(ctx context.Context, sessionID, method string, identityID *string, authenticatedAt time.Time) error
	GetSession(ctx context.Context, sessionID string) (identity.AuthSession, error)
}

type OIDCClient interface {
	AuthorizationURL(ctx context.Context, state, nonce, pkceVerifier string, forceReauthentication bool) (string, error)
	Exchange(ctx context.Context, code, pkceVerifier, nonce string) (OIDCClaims, error)
}

type OIDCFlowTokenManager interface {
	Generate(ctx context.Context) (string, error)
	Hash(value string) string
}

type APITokenRevoker interface {
	RevokeUserTokens(ctx context.Context, userID, reason string) error
}

type PasswordResetTokenManager interface {
	Generate(ctx context.Context) (string, error)
	Hash(value string) string
}

type EmailVerificationTokenManager interface {
	Generate(ctx context.Context) (string, error)
	Hash(value string) string
}

type PasswordResetMailer interface {
	SendPasswordReset(ctx context.Context, input identity.PasswordResetEmail) error
}

type EmailVerificationMailer interface {
	SendEmailVerification(ctx context.Context, input identity.EmailVerificationEmail) error
}

// SecurityEventRecorder records security-relevant auth events.
type SecurityEventRecorder interface {
	RecordAuthEvent(ctx context.Context, event AuthEvent)
}

type AuthEventName string

const (
	AuthEventRegisterSuccess                 AuthEventName = "auth.register.success"
	AuthEventRegisterFailure                 AuthEventName = "auth.register.failure"
	AuthEventLoginSuccess                    AuthEventName = "auth.login.success"
	AuthEventLoginFailure                    AuthEventName = "auth.login.failure"
	AuthEventSessionCreateFailure            AuthEventName = "auth.session.create.failure"
	AuthEventResetRequestSuccess             AuthEventName = "auth.password_reset.request.success"
	AuthEventResetRequestFailure             AuthEventName = "auth.password_reset.request.failure"
	AuthEventResetConfirmSuccess             AuthEventName = "auth.password_reset.confirm.success"
	AuthEventResetConfirmFailure             AuthEventName = "auth.password_reset.confirm.failure"
	AuthEventEmailVerificationRequestSuccess AuthEventName = "auth.email_verification.request.success"
	AuthEventEmailVerificationRequestFailure AuthEventName = "auth.email_verification.request.failure"
	AuthEventEmailVerificationConfirmSuccess AuthEventName = "auth.email_verification.confirm.success"
	AuthEventEmailVerificationConfirmFailure AuthEventName = "auth.email_verification.confirm.failure"
)

type AuthEventAction string

const (
	AuthActionRegister                 AuthEventAction = "register"
	AuthActionLogin                    AuthEventAction = "login"
	AuthActionResetRequest             AuthEventAction = "password_reset.request"
	AuthActionResetConfirm             AuthEventAction = "password_reset.confirm"
	AuthActionEmailVerificationRequest AuthEventAction = "email_verification.request"
	AuthActionEmailVerificationConfirm AuthEventAction = "email_verification.confirm"
)

type AuthEventOutcome string

const (
	AuthOutcomeSuccess AuthEventOutcome = "success"
	AuthOutcomeFailure AuthEventOutcome = "failure"
)

type AuthEventReason string

const (
	AuthReasonCredentialsInvalid                 AuthEventReason = "credentials_invalid"
	AuthReasonEmailAlreadyExists                 AuthEventReason = "email_already_exists"
	AuthReasonInvalidInput                       AuthEventReason = "invalid_input"
	AuthReasonPasswordHashFailed                 AuthEventReason = "password_hash_failed"
	AuthReasonResetMailerFailed                  AuthEventReason = "password_reset_mail_failed"
	AuthReasonResetTokenCreateFail               AuthEventReason = "password_reset_token_create_failed"
	AuthReasonResetTokenGenerateFail             AuthEventReason = "password_reset_token_generate_failed"
	AuthReasonResetTokenInvalid                  AuthEventReason = "password_reset_token_invalid"
	AuthReasonResetUnavailable                   AuthEventReason = "password_reset_unavailable"
	AuthReasonPasswordResetFailed                AuthEventReason = "password_reset_failed"
	AuthReasonEmailVerificationRequired          AuthEventReason = "email_verification_required"
	AuthReasonEmailVerificationUnavailable       AuthEventReason = "email_verification_unavailable"
	AuthReasonEmailVerificationTokenCreateFail   AuthEventReason = "email_verification_token_create_failed"   //nolint:gosec // Event reason names are not credentials.
	AuthReasonEmailVerificationTokenGenerateFail AuthEventReason = "email_verification_token_generate_failed" //nolint:gosec // Event reason names are not credentials.
	AuthReasonEmailVerificationTokenInvalid      AuthEventReason = "email_verification_token_invalid"         //nolint:gosec // Event reason names are not credentials.
	AuthReasonEmailVerificationMailerFailed      AuthEventReason = "email_verification_mail_failed"
	AuthReasonEmailVerificationFailed            AuthEventReason = "email_verification_failed"
	AuthReasonAccountDisabled                    AuthEventReason = "account_disabled"
	AuthReasonUserCreateFailed                   AuthEventReason = "user_create_failed"
	AuthReasonUserLookupFailed                   AuthEventReason = "user_lookup_failed"
	AuthReasonSessionCreateFail                  AuthEventReason = "session_create_failed"
	AuthReasonSessionListFail                    AuthEventReason = "session_list_failed"
	AuthReasonSessionRevokeFail                  AuthEventReason = "session_revoke_failed"
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
