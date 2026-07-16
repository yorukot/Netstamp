package apitoken

import (
	"context"
	"errors"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

const touchInterval = 5 * time.Minute

type Service struct {
	repo   Repository
	tokens TokenManager
	events EventRecorder
	now    func() time.Time
}

func NewService(repo Repository, tokens TokenManager, events EventRecorder) *Service {
	return &Service{repo: repo, tokens: tokens, events: events, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (CreateOutput, error) {
	ctx, flow := startFlow(ctx, s.events, "api_token.create", input.CurrentUserID)
	defer flow.end()
	now := s.now()
	normalized, err := normalizeCreateInput(input, now)
	if err != nil {
		flow.record("api_token.create.failure", "failure", "invalid_input", "", err)
		return CreateOutput{}, err
	}
	raw, hint, err := s.tokens.Generate()
	if err != nil {
		flow.record("api_token.create.failure", "failure", "token_generate_failed", "", err)
		return CreateOutput{}, err
	}
	scopes := make([]identity.APITokenScope, 0, len(normalized.Scopes))
	for _, scope := range normalized.Scopes {
		scopes = append(scopes, identity.APITokenScope(scope))
	}
	token, err := s.repo.Create(ctx, identity.APIToken{UserID: normalized.CurrentUserID, Name: normalized.Name, TokenHash: s.tokens.Hash(raw), TokenHint: hint, Scopes: scopes, CreatedAt: now, ExpiresAt: normalized.ExpiresAt}, MaxActiveTokens, now)
	if err != nil {
		if errors.Is(err, identity.ErrAPITokenLimitReached) {
			err = ErrTokenLimitReached
		}
		flow.record("api_token.create.failure", "failure", "token_create_failed", "", err)
		return CreateOutput{}, err
	}
	flow.record("api_token.create.success", "success", "", token.ID, nil)
	return CreateOutput{Token: token, RawToken: raw}, nil
}

func (s *Service) List(ctx context.Context, input ListInput) ([]identity.APIToken, error) {
	if input.CurrentUserID == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.ListForUser(ctx, input.CurrentUserID)
}

func (s *Service) Revoke(ctx context.Context, input RevokeInput) error {
	ctx, flow := startFlow(ctx, s.events, "api_token.revoke", input.CurrentUserID)
	defer flow.end()
	if input.CurrentUserID == "" || input.TokenID == "" {
		return ErrInvalidInput
	}
	if err := s.repo.RevokeForUser(ctx, input.CurrentUserID, input.TokenID, "user_revoked", s.now()); err != nil {
		if errors.Is(err, identity.ErrAPITokenNotFound) {
			err = ErrTokenNotFound
		}
		flow.record("api_token.revoke.failure", "failure", "token_revoke_failed", input.TokenID, err)
		return err
	}
	flow.record("api_token.revoke.success", "success", "", input.TokenID, nil)
	return nil
}

func (s *Service) Verify(ctx context.Context, rawToken string) (Principal, error) {
	if rawToken == "" {
		return Principal{}, ErrTokenInvalid
	}
	now := s.now()
	token, err := s.repo.GetActiveByHash(ctx, s.tokens.Hash(rawToken), now)
	if err != nil {
		if errors.Is(err, identity.ErrAPITokenNotFound) {
			return Principal{}, ErrTokenInvalid
		}
		return Principal{}, err
	}
	if token.LastUsedAt == nil || now.Sub(*token.LastUsedAt) >= touchInterval {
		if err := s.repo.Touch(ctx, token.ID, now, now.Add(-touchInterval)); err != nil {
			return Principal{}, err
		}
	}
	return Principal{TokenID: token.ID, UserID: token.UserID, Scopes: token.Scopes}, nil
}

func (s *Service) RevokeUserTokens(ctx context.Context, userID, reason string) error {
	if userID == "" {
		return nil
	}
	return s.repo.RevokeForUserAll(ctx, userID, reason, s.now())
}
