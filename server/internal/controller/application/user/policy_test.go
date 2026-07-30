package account

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

const (
	policyTestUserID     = "11111111-1111-1111-1111-111111111111"
	policyTestSessionID  = "22222222-2222-2222-2222-222222222222"
	policyTestIdentityID = "33333333-3333-3333-3333-333333333333"
)

func TestCredentialChangesPolicyEnabledAllowsCredentialMutations(t *testing.T) {
	tests := credentialPolicyOperations()

	for name, operation := range tests {
		t.Run(name, func(t *testing.T) {
			dependencies := &policyUserDependencies{}
			policy := &credentialChangesPolicyStub{enabled: true}
			service := newPolicyUserService(dependencies, &policyUserEventRecorder{}, policy)

			if err := operation(service); err != nil {
				t.Fatalf("expected credential mutation to succeed: %v", err)
			}
			if policy.calls != 1 {
				t.Fatalf("expected one policy lookup, got %d", policy.calls)
			}
			if dependencies.calls == 0 {
				t.Fatalf("expected credential mutation to reach downstream dependencies")
			}
		})
	}
}

func TestCredentialChangesPolicyDisabledBlocksCredentialMutations(t *testing.T) {
	tests := credentialPolicyOperations()

	for name, operation := range tests {
		t.Run(name, func(t *testing.T) {
			dependencies := &policyUserDependencies{}
			events := &policyUserEventRecorder{}
			policy := &credentialChangesPolicyStub{enabled: false}
			service := newPolicyUserService(dependencies, events, policy)

			err := operation(service)

			if !errors.Is(err, ErrForbidden) {
				t.Fatalf("expected credential changes to be forbidden, got %v", err)
			}
			if policy.calls != 1 {
				t.Fatalf("expected one policy lookup, got %d", policy.calls)
			}
			if dependencies.calls != 0 {
				t.Fatalf("expected policy denial before downstream work, got %d calls", dependencies.calls)
			}
		})
	}
}

func TestCredentialChangesPolicyErrorsFailAsTechnicalErrors(t *testing.T) {
	policyErr := errors.New("settings unavailable")
	tests := credentialPolicyOperations()

	for name, operation := range tests {
		t.Run(name, func(t *testing.T) {
			dependencies := &policyUserDependencies{}
			events := &policyUserEventRecorder{}
			policy := &credentialChangesPolicyStub{err: policyErr}
			service := newPolicyUserService(dependencies, events, policy)

			err := operation(service)

			if !errors.Is(err, policyErr) {
				t.Fatalf("expected policy lookup error, got %v", err)
			}
			if errors.Is(err, ErrForbidden) {
				t.Fatalf("expected policy lookup failure not to be reported as forbidden")
			}
			if policy.calls != 1 {
				t.Fatalf("expected one policy lookup, got %d", policy.calls)
			}
			if dependencies.calls != 0 {
				t.Fatalf("expected policy failure before downstream work, got %d calls", dependencies.calls)
			}

			if name == "change email" || name == "change password" {
				if len(events.events) != 1 {
					t.Fatalf("expected one failure event, got %d", len(events.events))
				}
				event := events.events[0]
				if event.Reason != UserReasonPolicyLookupFailed {
					t.Fatalf("expected policy lookup failure reason, got %q", event.Reason)
				}
				if !errors.Is(event.Err, policyErr) {
					t.Fatalf("expected technical event to retain policy error, got %v", event.Err)
				}
			}
		})
	}
}

func credentialPolicyOperations() map[string]func(*Service) error {
	return map[string]func(*Service) error{
		"change email": func(service *Service) error {
			_, err := service.ChangeCurrentUserEmail(context.Background(), ChangeCurrentUserEmailInput{
				CurrentUserID: policyTestUserID,
				NewEmail:      "changed@example.com",
			})
			return err
		},
		"change password": func(service *Service) error {
			return service.ChangeCurrentUserPassword(context.Background(), ChangeCurrentUserPasswordInput{
				CurrentUserID:    policyTestUserID,
				CurrentSessionID: policyTestSessionID,
				NewPassword:      "new-password",
			})
		},
		"remove password": func(service *Service) error {
			return service.RemoveCurrentUserPassword(context.Background(), policyTestUserID, policyTestSessionID)
		},
		"remove identity": func(service *Service) error {
			return service.RemoveCurrentUserIdentity(
				context.Background(),
				policyTestUserID,
				policyTestSessionID,
				policyTestIdentityID,
			)
		},
	}
}

func newPolicyUserService(
	dependencies *policyUserDependencies,
	events *policyUserEventRecorder,
	policy CredentialChangesPolicy,
) *Service {
	service := NewService(dependencies, dependencies, events)
	service.ConfigureAuthenticationMethods(dependencies)
	service.ConfigureInstancePolicy(policy)
	return service
}

type credentialChangesPolicyStub struct {
	enabled bool
	err     error
	calls   int
}

func (p *credentialChangesPolicyStub) CredentialChangesEnabled(context.Context) (bool, error) {
	p.calls++
	return p.enabled, p.err
}

type policyUserDependencies struct {
	calls int
}

func (d *policyUserDependencies) GetUserByID(context.Context, string) (identity.User, error) {
	d.calls++
	return identity.User{ID: policyTestUserID, Email: "user@example.com"}, nil
}

func (d *policyUserDependencies) UpdateUserDisplayName(_ context.Context, user identity.User) (identity.User, error) {
	d.calls++
	return user, nil
}

func (d *policyUserDependencies) UpdateUserEmail(_ context.Context, user identity.User) (identity.User, error) {
	d.calls++
	return user, nil
}

func (d *policyUserDependencies) UpdateUserPasswordHash(_ context.Context, user identity.User) (identity.User, error) {
	d.calls++
	return user, nil
}

func (d *policyUserDependencies) DisableUser(_ context.Context, userID string) (identity.User, error) {
	d.calls++
	return identity.User{ID: userID}, nil
}

func (d *policyUserDependencies) Hash(context.Context, string) (string, error) {
	d.calls++
	return "password-hash", nil
}

func (d *policyUserDependencies) ListUserIdentities(context.Context, string) ([]identity.UserIdentity, error) {
	d.calls++
	return nil, nil
}

func (d *policyUserDependencies) CountUserAuthenticationMethods(context.Context, string) (bool, int64, error) {
	d.calls++
	return true, 1, nil
}

func (d *policyUserDependencies) DeleteUserIdentity(context.Context, string, string) error {
	d.calls++
	return nil
}

func (d *policyUserDependencies) DeleteUserPasswordCredential(_ context.Context, userID string) (identity.User, error) {
	d.calls++
	return identity.User{ID: userID}, nil
}

type policyUserEventRecorder struct {
	events []UserEvent
}

func (r *policyUserEventRecorder) RecordUserEvent(_ context.Context, event UserEvent) {
	r.events = append(r.events, event)
}
