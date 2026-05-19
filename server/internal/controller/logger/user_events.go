package logger

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"go.uber.org/zap"

	appuser "github.com/yorukot/netstamp/internal/controller/application/user"
)

type UserEventRecorder struct {
	root         *zap.Logger
	pseudonymKey []byte
}

func NewUserEventRecorder(root *zap.Logger, pseudonymKey string) *UserEventRecorder {
	if root == nil {
		root = zap.NewNop()
	}

	return &UserEventRecorder{
		root:         root,
		pseudonymKey: []byte(pseudonymKey),
	}
}

func (r *UserEventRecorder) RecordUserEvent(ctx context.Context, event appuser.UserEvent) {
	fields := make([]zap.Field, 0, 2)
	if event.UserID != "" {
		fields = append(fields, zap.String("user.id", event.UserID))
	}
	if emailHash := r.emailHash(event.Email); emailHash != "" {
		fields = append(fields, zap.String("user.email_hash", emailHash))
	}
	recordApplicationEvent(ctx, r.root, applicationEventLog{
		name:            string(event.Name),
		category:        "user",
		action:          string(event.Action),
		outcome:         string(event.Outcome),
		reason:          string(event.Reason),
		successOutcome:  string(appuser.UserOutcomeSuccess),
		expectedFailure: isExpectedUserFailure(event),
		fields:          fields,
		err:             event.Err,
	})
}

func (r *UserEventRecorder) emailHash(email string) string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return ""
	}

	mac := hmac.New(sha256.New, r.pseudonymKey)
	_, _ = mac.Write([]byte(normalized))
	return hex.EncodeToString(mac.Sum(nil))
}

func isExpectedUserFailure(event appuser.UserEvent) bool {
	switch event.Reason {
	case appuser.UserReasonInvalidInput,
		appuser.UserReasonCredentialsInvalid,
		appuser.UserReasonEmailAlreadyExists,
		appuser.UserReasonUserNotFound:
		return true
	default:
		return false
	}
}
