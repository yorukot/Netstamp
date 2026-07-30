package auth

import (
	"context"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

func (s *Service) SudoStatus(ctx context.Context, userID, sessionID string) (SudoStatusResult, error) {
	if s.recentAuth == nil {
		return SudoStatusResult{}, ErrSessionInvalid
	}
	status, err := s.recentAuth.SudoStatus(ctx, sessionID)
	if err != nil {
		return SudoStatusResult{}, err
	}
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return SudoStatusResult{}, err
	}
	methods := make([]string, 0, 3)
	if user.HasPassword {
		methods = append(methods, identity.AuthenticationMethodPassword)
	}
	if s.externalAuthRepo != nil {
		identities, err := s.externalAuthRepo.ListUserIdentities(ctx, userID)
		if err != nil {
			return SudoStatusResult{}, err
		}
		seen := make(map[string]bool)
		for _, linked := range identities {
			providerID := linked.Provider
			if seen[providerID] {
				continue
			}
			seen[providerID] = true
			provider, resolveErr := s.resolvedExternalProvider(ctx, providerID)
			if resolveErr == nil && provider.config.SudoCapable {
				methods = append(methods, providerID)
			}
		}
	}
	return SudoStatusResult{Active: status.Active, ExpiresAt: status.ExpiresAt, Methods: methods}, nil
}

func (s *Service) ReauthenticatePassword(ctx context.Context, userID, sessionID, password string) error {
	if userID == "" || sessionID == "" || password == "" {
		return ErrCredentialsInvalid
	}
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil || !user.HasPassword {
		return ErrCredentialsInvalid
	}
	if err := s.comparePassword(ctx, password, user.PasswordHash); err != nil {
		return ErrCredentialsInvalid
	}
	if s.recentAuth == nil {
		return ErrSessionInvalid
	}
	return s.recentAuth.ElevateSession(ctx, sessionID, identity.AuthenticationMethodPassword, nil, s.now())
}

func (s *Service) RequireSudo(ctx context.Context, sessionID string) error {
	if s.recentAuth == nil {
		return ErrSessionInvalid
	}
	status, err := s.recentAuth.SudoStatus(ctx, sessionID)
	if err != nil {
		return err
	}
	if !status.Active {
		return ErrSudoRequired
	}
	return nil
}
