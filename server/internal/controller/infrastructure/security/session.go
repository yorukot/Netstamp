package security

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"time"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

const sessionTokenBytes = 32

type SessionConfig struct {
	HashKey       string
	IdleTTL       time.Duration
	AbsoluteTTL   time.Duration
	TouchInterval time.Duration
	SudoTTL       time.Duration
}

type SessionRepository interface {
	CreateSession(ctx context.Context, input identity.AuthSession) (identity.AuthSession, error)
	GetActiveSessionByTokenHash(ctx context.Context, tokenHash []byte, now time.Time) (identity.AuthSession, error)
	GetActiveSessionByID(ctx context.Context, sessionID string, now time.Time) (identity.AuthSession, error)
	UpdateCSRFTokenHash(ctx context.Context, sessionID string, csrfTokenHash []byte, now time.Time) error
	TouchSession(ctx context.Context, sessionID string, lastUsedAt, idleExpiresAt time.Time) error
	RevokeSessionByTokenHash(ctx context.Context, tokenHash []byte, revokedAt time.Time, reason string) error
	ListActiveSessionsForUser(ctx context.Context, userID string, now time.Time) ([]identity.AuthSession, error)
	RevokeSessionByIDForUser(ctx context.Context, userID, sessionID string, revokedAt time.Time, reason string) error
	RevokeSessionsForUser(ctx context.Context, userID string, revokedAt time.Time, reason string) error
}

type RecentAuthenticationRepository interface {
	UpdateSessionAuthentication(ctx context.Context, sessionID string, authenticatedAt time.Time, method string, identityID *string) error
}

type OtherSessionRepository interface {
	RevokeSessionsForUserExcept(ctx context.Context, userID, sessionID string, revokedAt time.Time, reason string) error
}

type SessionManager struct {
	repo          SessionRepository
	hashKey       []byte
	idleTTL       time.Duration
	absoluteTTL   time.Duration
	touchInterval time.Duration
	sudoTTL       time.Duration
	now           func() time.Time
}

func NewSessionManager(repo SessionRepository, cfg SessionConfig) *SessionManager {
	return &SessionManager{
		repo:          repo,
		hashKey:       []byte(cfg.HashKey),
		idleTTL:       cfg.IdleTTL,
		absoluteTTL:   cfg.AbsoluteTTL,
		touchInterval: cfg.TouchInterval,
		sudoTTL:       cfg.SudoTTL,
		now:           func() time.Time { return time.Now().UTC() },
	}
}

func (m *SessionManager) CreateSession(ctx context.Context, input appauth.CreateSessionInput) (identity.CreatedSession, error) {
	if m.repo == nil || len(m.hashKey) == 0 {
		return identity.CreatedSession{}, appauth.ErrSessionInvalid
	}

	now := input.Now.UTC()
	if now.IsZero() {
		now = m.now()
	}

	rawToken, err := randomToken()
	if err != nil {
		return identity.CreatedSession{}, err
	}
	rawCSRFToken, err := randomToken()
	if err != nil {
		return identity.CreatedSession{}, err
	}

	idleExpiresAt := now.Add(m.idleTTL)
	absoluteExpiresAt := now.Add(m.absoluteTTL)
	if idleExpiresAt.After(absoluteExpiresAt) {
		idleExpiresAt = absoluteExpiresAt
	}

	method := authenticationMethod(input.AuthenticationMethod)
	session, err := m.repo.CreateSession(ctx, identity.AuthSession{
		UserID:               input.UserID,
		TokenHash:            m.hash(rawToken),
		CSRFTokenHash:        m.hash(rawCSRFToken),
		UserAgent:            input.UserAgent,
		AuthenticatedAt:      now,
		AuthenticationMethod: method,
		SudoEligible:         input.SudoEligible && authenticationMethodSupportsSudo(method),
		IdentityID:           input.IdentityID,
		CreatedAt:            now,
		LastUsedAt:           now,
		IdleExpiresAt:        idleExpiresAt,
		AbsoluteExpiresAt:    absoluteExpiresAt,
	})
	if err != nil {
		return identity.CreatedSession{}, err
	}

	return identity.CreatedSession{
		Session:         session,
		RawToken:        rawToken,
		RawCSRFToken:    rawCSRFToken,
		ExpiresIn:       int(absoluteExpiresAt.Sub(now).Seconds()),
		CookieExpiresAt: absoluteExpiresAt,
	}, nil
}

func authenticationMethod(method string) string {
	switch method {
	case identity.AuthenticationMethodGoogle, identity.AuthenticationMethodGitHub, identity.AuthenticationMethodOIDC:
		return method
	default:
		return identity.AuthenticationMethodPassword
	}
}

func (m *SessionManager) SudoStatus(ctx context.Context, sessionID string) (identity.SudoStatus, error) {
	session, err := m.activeSessionByID(ctx, sessionID)
	if err != nil {
		return identity.SudoStatus{}, err
	}
	expiresAt := session.AuthenticatedAt.Add(m.sudoTTL)
	return identity.SudoStatus{Active: session.SudoEligible && authenticationMethodSupportsSudo(session.AuthenticationMethod) && m.now().Before(expiresAt), ExpiresAt: expiresAt}, nil
}

func authenticationMethodSupportsSudo(method string) bool {
	return method == identity.AuthenticationMethodPassword || method == identity.AuthenticationMethodGoogle || method == identity.AuthenticationMethodOIDC
}

func (m *SessionManager) GetSession(ctx context.Context, sessionID string) (identity.AuthSession, error) {
	return m.activeSessionByID(ctx, sessionID)
}

func (m *SessionManager) ElevateSession(ctx context.Context, sessionID, method string, identityID *string, authenticatedAt time.Time) error {
	if sessionID == "" || m.sudoTTL <= 0 {
		return appauth.ErrSessionInvalid
	}
	if method != identity.AuthenticationMethodPassword && method != identity.AuthenticationMethodGoogle && method != identity.AuthenticationMethodOIDC {
		return appauth.ErrSessionInvalid
	}
	if authenticatedAt.IsZero() {
		authenticatedAt = m.now()
	}
	repo, ok := m.repo.(RecentAuthenticationRepository)
	if !ok {
		return appauth.ErrSessionInvalid
	}
	if err := repo.UpdateSessionAuthentication(ctx, sessionID, authenticatedAt.UTC(), method, identityID); err != nil {
		if errors.Is(err, identity.ErrSessionNotFound) {
			return appauth.ErrSessionInvalid
		}
		return err
	}
	return nil
}

func (m *SessionManager) VerifySession(ctx context.Context, rawToken string) (identity.SessionClaims, error) {
	session, err := m.activeSession(ctx, rawToken)
	if err != nil {
		return identity.SessionClaims{}, err
	}

	now := m.now()
	if now.Sub(session.LastUsedAt) >= m.touchInterval {
		idleExpiresAt := now.Add(m.idleTTL)
		if idleExpiresAt.After(session.AbsoluteExpiresAt) {
			idleExpiresAt = session.AbsoluteExpiresAt
		}
		if err := m.repo.TouchSession(ctx, session.ID, now, idleExpiresAt); err != nil {
			return identity.SessionClaims{}, err
		}
	}

	return identity.SessionClaims{SessionID: session.ID, UserID: session.UserID}, nil
}

func (m *SessionManager) CreateCSRFToken(ctx context.Context, sessionID string) (string, error) {
	if sessionID == "" {
		return "", appauth.ErrSessionInvalid
	}
	rawCSRFToken, err := randomToken()
	if err != nil {
		return "", err
	}
	if err := m.repo.UpdateCSRFTokenHash(ctx, sessionID, m.hash(rawCSRFToken), m.now()); err != nil {
		if errors.Is(err, identity.ErrSessionNotFound) {
			return "", appauth.ErrSessionInvalid
		}
		return "", err
	}
	return rawCSRFToken, nil
}

func (m *SessionManager) VerifyCSRFToken(ctx context.Context, sessionID, rawToken string) error {
	if rawToken == "" {
		return appauth.ErrSessionInvalid
	}
	session, err := m.activeSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}
	expected := session.CSRFTokenHash
	actual := m.hash(rawToken)
	if subtle.ConstantTimeCompare(expected, actual) != 1 {
		return appauth.ErrSessionInvalid
	}
	return nil
}

func (m *SessionManager) RevokeSession(ctx context.Context, rawToken, reason string) error {
	if rawToken == "" {
		return nil
	}
	if reason == "" {
		reason = "logout"
	}
	return m.repo.RevokeSessionByTokenHash(ctx, m.hash(rawToken), m.now(), reason)
}

func (m *SessionManager) ListUserSessions(ctx context.Context, userID string) ([]identity.AuthSession, error) {
	if userID == "" {
		return nil, appauth.ErrSessionInvalid
	}
	return m.repo.ListActiveSessionsForUser(ctx, userID, m.now())
}

func (m *SessionManager) RevokeUserSession(ctx context.Context, userID, sessionID, reason string) error {
	if userID == "" || sessionID == "" {
		return identity.ErrSessionNotFound
	}
	if reason == "" {
		reason = "user_revoked"
	}
	return m.repo.RevokeSessionByIDForUser(ctx, userID, sessionID, m.now(), reason)
}

func (m *SessionManager) RevokeUserSessions(ctx context.Context, userID, reason string) error {
	if userID == "" {
		return nil
	}
	if reason == "" {
		reason = "security_event"
	}
	return m.repo.RevokeSessionsForUser(ctx, userID, m.now(), reason)
}

func (m *SessionManager) RevokeUserSessionsExcept(ctx context.Context, userID, sessionID, reason string) error {
	if userID == "" || sessionID == "" {
		return appauth.ErrSessionInvalid
	}
	if reason == "" {
		reason = "security_event"
	}
	repo, ok := m.repo.(OtherSessionRepository)
	if !ok {
		return appauth.ErrSessionInvalid
	}
	return repo.RevokeSessionsForUserExcept(ctx, userID, sessionID, m.now(), reason)
}

func (m *SessionManager) activeSession(ctx context.Context, rawToken string) (identity.AuthSession, error) {
	if rawToken == "" {
		return identity.AuthSession{}, appauth.ErrSessionInvalid
	}
	session, err := m.repo.GetActiveSessionByTokenHash(ctx, m.hash(rawToken), m.now())
	if err != nil {
		if errors.Is(err, identity.ErrSessionNotFound) {
			return identity.AuthSession{}, appauth.ErrSessionInvalid
		}
		return identity.AuthSession{}, err
	}
	return session, nil
}

func (m *SessionManager) activeSessionByID(ctx context.Context, sessionID string) (identity.AuthSession, error) {
	if sessionID == "" {
		return identity.AuthSession{}, appauth.ErrSessionInvalid
	}
	session, err := m.repo.GetActiveSessionByID(ctx, sessionID, m.now())
	if err != nil {
		if errors.Is(err, identity.ErrSessionNotFound) {
			return identity.AuthSession{}, appauth.ErrSessionInvalid
		}
		return identity.AuthSession{}, err
	}
	return session, nil
}

func (m *SessionManager) hash(value string) []byte {
	mac := hmac.New(sha256.New, m.hashKey)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}

func randomToken() (string, error) {
	var token [sessionTokenBytes]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(token[:]), nil
}
