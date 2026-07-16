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
	methods := make([]string, 0, 2)
	if user.HasPassword {
		methods = append(methods, identity.AuthenticationMethodPassword)
	}
	if s.oidcRepo != nil {
		identities, err := s.oidcRepo.ListUserIdentities(ctx, userID)
		if err != nil {
			return SudoStatusResult{}, err
		}
		if len(identities) > 0 && s.oidcConfig.Enabled {
			methods = append(methods, identity.AuthenticationMethodOIDC)
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
