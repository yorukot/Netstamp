package logger

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"go.uber.org/zap"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
)

type AuthEventRecorder struct {
	root         *zap.Logger
	pseudonymKey []byte
}

func NewAuthEventRecorder(root *zap.Logger, pseudonymKey string) *AuthEventRecorder {
	if root == nil {
		root = zap.NewNop()
	}

	return &AuthEventRecorder{
		root:         root,
		pseudonymKey: []byte(pseudonymKey),
	}
}

func (r *AuthEventRecorder) RecordAuthEvent(ctx context.Context, event appauth.AuthEvent) {
	log := FromContext(ctx, r.root)
	fields := eventFields(string(event.Name), "auth", string(event.Action), string(event.Outcome), string(event.Reason))

	if event.UserID != "" {
		fields = append(fields, zap.String("user.id", event.UserID))
	}
	if emailHash := r.emailHash(event.Email); emailHash != "" {
		fields = append(fields, zap.String("user.email_hash", emailHash))
	}
	if event.Err != nil {
		fields = append(fields, zap.Error(event.Err))
	}

	switch {
	case event.Outcome == appauth.AuthOutcomeSuccess:
		log.Info(string(event.Name), fields...)
	case isExpectedAuthFailure(event):
		log.Warn(string(event.Name), fields...)
	default:
		log.Error(string(event.Name), fields...)
	}
}

func (r *AuthEventRecorder) emailHash(email string) string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return ""
	}

	mac := hmac.New(sha256.New, r.pseudonymKey)
	_, _ = mac.Write([]byte(normalized))
	return hex.EncodeToString(mac.Sum(nil))
}

func isExpectedAuthFailure(event appauth.AuthEvent) bool {
	switch event.Reason {
	case appauth.AuthReasonCredentialsInvalid, appauth.AuthReasonEmailAlreadyExists, appauth.AuthReasonUserInactive:
		return true
	default:
		return false
	}
}
