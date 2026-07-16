package auth

import (
	"context"
	"errors"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

func (s *Service) AuthorizePasswordChange(ctx context.Context, userID, sessionID string) error {
	if userID == "" || sessionID == "" || s.recentAuth == nil {
		return ErrSessionInvalid
	}
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.HasPassword {
		return s.RequireSudo(ctx, sessionID)
	}
	if s.externalAuthRepo == nil {
		return ErrSudoRequired
	}
	session, err := s.recentAuth.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if session.UserID != userID || session.AuthenticationMethod != identity.AuthenticationMethodGitHub || session.IdentityID == nil {
		return ErrSudoRequired
	}
	linked, err := s.externalAuthRepo.GetUserIdentityByIDForUser(ctx, *session.IdentityID, userID)
	if err != nil {
		if errors.Is(err, identity.ErrIdentityNotFound) {
			return ErrSudoRequired
		}
		return err
	}
	if linked.Provider != identity.AuthenticationMethodGitHub {
		return ErrSudoRequired
	}
	now := s.now()
	if session.CreatedAt.IsZero() || session.CreatedAt.After(now.Add(s.externalAuthConfig.AuthTimeSkew)) || now.Sub(session.CreatedAt) > s.externalAuthConfig.FlowTTL {
		return ErrSudoRequired
	}
	return nil
}

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
			provider, enabled := s.externalProviders[linked.Provider]
			if enabled && provider.config.SudoCapable && !seen[linked.Provider] {
				methods = append(methods, linked.Provider)
				seen[linked.Provider] = true
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
