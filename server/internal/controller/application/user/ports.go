package account

import (
	"context"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

type Repository interface {
	GetUserByID(ctx context.Context, userID string) (identity.User, error)
	UpdateUserDisplayName(ctx context.Context, input identity.User) (identity.User, error)
	UpdateUserEmail(ctx context.Context, input identity.User) (identity.User, error)
	UpdateUserPasswordHash(ctx context.Context, input identity.User) (identity.User, error)
	DisableUser(ctx context.Context, userID string) (identity.User, error)
}

type SystemAdminRepository interface {
	CountActiveSystemAdmins(ctx context.Context) (int64, error)
}

type SessionRepository interface {
	RevokeUserSessions(ctx context.Context, userID, reason string) error
}

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Compare(ctx context.Context, password, passwordHash string) error
}

type EventRecorder interface {
	RecordUserEvent(ctx context.Context, event UserEvent)
}

type UserEventName string

const (
	UserEventUpdateProfileSuccess  UserEventName = "user.profile.update.success"
	UserEventUpdateProfileFailure  UserEventName = "user.profile.update.failure"
	UserEventChangeEmailSuccess    UserEventName = "user.email.change.success"
	UserEventChangeEmailFailure    UserEventName = "user.email.change.failure"
	UserEventChangePasswordSuccess UserEventName = "user.password.change.success"
	UserEventChangePasswordFailure UserEventName = "user.password.change.failure"
	UserEventDeactivateSuccess     UserEventName = "user.deactivate.success"
	UserEventDeactivateFailure     UserEventName = "user.deactivate.failure"
)

type UserEventAction string

const (
	UserActionUpdateProfile  UserEventAction = "profile.update"
	UserActionChangeEmail    UserEventAction = "email.change"
	UserActionChangePassword UserEventAction = "password.change"
	UserActionDeactivate     UserEventAction = "deactivate"
)

type UserEventOutcome string

const (
	UserOutcomeSuccess UserEventOutcome = "success"
	UserOutcomeFailure UserEventOutcome = "failure"
)

type UserEventReason string

const (
	UserReasonInvalidInput       UserEventReason = "invalid_input"
	UserReasonCredentialsInvalid UserEventReason = "credentials_invalid"
	UserReasonEmailAlreadyExists UserEventReason = "email_already_exists"
	UserReasonUserNotFound       UserEventReason = "user_not_found"
	UserReasonUserLookupFailed   UserEventReason = "user_lookup_failed"
	UserReasonUserUpdateFailed   UserEventReason = "user_update_failed"
	UserReasonPasswordHashFailed UserEventReason = "password_hash_failed"
)

type UserEvent struct {
	Name    UserEventName
	Action  UserEventAction
	Outcome UserEventOutcome
	Reason  UserEventReason
	UserID  string
	Email   string
	Err     error
}
