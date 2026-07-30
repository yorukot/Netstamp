package admin

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

const (
	testAdminUserID  = "11111111-1111-1111-1111-111111111111"
	testTargetUserID = "22222222-2222-2222-2222-222222222222"
)

func TestGrantSystemAdminStoresAuditEvent(t *testing.T) {
	grantedAt := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	repo := &fakeAdminRepository{
		admins: map[string]bool{testAdminUserID: true},
		systemAdminByEmail: map[string]SystemAdmin{
			"operator@example.com": {
				ID:          testTargetUserID,
				Email:       "operator@example.com",
				DisplayName: "Operator",
				GrantedAt:   grantedAt,
			},
		},
	}
	svc := NewService(repo)

	admin, err := svc.GrantSystemAdmin(context.Background(), GrantSystemAdminInput{
		CurrentUserID: testAdminUserID,
		Email:         " OPERATOR@example.com ",
	})
	if err != nil {
		t.Fatalf("grant system admin: %v", err)
	}

	if admin.ID != testTargetUserID {
		t.Fatalf("expected granted admin %q, got %q", testTargetUserID, admin.ID)
	}
	if repo.grantedEmail != "operator@example.com" {
		t.Fatalf("expected normalized grant email, got %q", repo.grantedEmail)
	}
	if !slices.Contains(repo.auditKeys, "system_admin:"+testTargetUserID) {
		t.Fatalf("expected system admin audit key, got %#v", repo.auditKeys)
	}
	if !slices.Contains(repo.auditActions, auditActionGrantSystemAdmin) {
		t.Fatalf("expected grant audit action, got %#v", repo.auditActions)
	}
}

func TestRevokeSystemAdminRejectsSelfRemoval(t *testing.T) {
	repo := &fakeAdminRepository{admins: map[string]bool{testAdminUserID: true}}
	svc := NewService(repo)

	err := svc.RevokeSystemAdmin(context.Background(), RevokeSystemAdminInput{
		CurrentUserID: testAdminUserID,
		UserID:        testAdminUserID,
	})
	if !errors.Is(err, ErrSelfSystemAdminRemoval) {
		t.Fatalf("expected self-removal error, got %v", err)
	}
	if repo.revokedUserID != "" {
		t.Fatalf("expected no revoke call, got %q", repo.revokedUserID)
	}
}

func TestRevokeSystemAdminRejectsLastAdmin(t *testing.T) {
	repo := &fakeAdminRepository{
		admins:       map[string]bool{testAdminUserID: true},
		revokeResult: SystemAdminRevokeResult{AdminCount: 1, TargetWasAdmin: true, Revoked: false},
	}
	svc := NewService(repo)

	err := svc.RevokeSystemAdmin(context.Background(), RevokeSystemAdminInput{
		CurrentUserID: testAdminUserID,
		UserID:        testTargetUserID,
	})
	if !errors.Is(err, ErrLastSystemAdmin) {
		t.Fatalf("expected last-admin error, got %v", err)
	}
}

func TestRevokeSystemAdminStoresAuditEvent(t *testing.T) {
	repo := &fakeAdminRepository{
		admins:       map[string]bool{testAdminUserID: true},
		revokeResult: SystemAdminRevokeResult{AdminCount: 2, TargetWasAdmin: true, Revoked: true},
	}
	svc := NewService(repo)

	err := svc.RevokeSystemAdmin(context.Background(), RevokeSystemAdminInput{
		CurrentUserID: testAdminUserID,
		UserID:        testTargetUserID,
	})
	if err != nil {
		t.Fatalf("revoke system admin: %v", err)
	}
	if repo.revokedUserID != testTargetUserID {
		t.Fatalf("expected revoked user %q, got %q", testTargetUserID, repo.revokedUserID)
	}
	if !slices.Contains(repo.auditKeys, "system_admin:"+testTargetUserID) {
		t.Fatalf("expected system admin audit key, got %#v", repo.auditKeys)
	}
	if !slices.Contains(repo.auditActions, auditActionRevokeSystemAdmin) {
		t.Fatalf("expected revoke audit action, got %#v", repo.auditActions)
	}
}

func TestClearManagedUserPasswordRequiresAnotherAuthenticationMethod(t *testing.T) {
	repo := &fakeAdminRepository{admins: map[string]bool{testAdminUserID: true}}
	svc := NewService(repo)
	svc.ConfigureAuthenticationMethods(fakeAuthenticationMethodRepository{identityCount: 0})

	_, err := svc.ClearManagedUserPassword(context.Background(), ClearManagedUserPasswordInput{CurrentUserID: testAdminUserID, UserID: testTargetUserID})
	if !errors.Is(err, identity.ErrLastAuthenticationMethod) {
		t.Fatalf("expected last authentication method error, got %v", err)
	}
	if repo.clearedPasswordUserID != "" {
		t.Fatalf("expected password to remain, clear called for %q", repo.clearedPasswordUserID)
	}
}

func TestClearManagedUserPasswordAuditsRemoval(t *testing.T) {
	repo := &fakeAdminRepository{
		admins: map[string]bool{testAdminUserID: true},
		managedUsers: []ManagedUser{{
			ID: testTargetUserID, Email: "person@example.com", DisplayName: "Person", HasPassword: true,
		}},
	}
	svc := NewService(repo)
	svc.ConfigureAuthenticationMethods(fakeAuthenticationMethodRepository{identityCount: 1})

	user, err := svc.ClearManagedUserPassword(context.Background(), ClearManagedUserPasswordInput{CurrentUserID: testAdminUserID, UserID: testTargetUserID})
	if err != nil {
		t.Fatalf("clear managed user password: %v", err)
	}
	if user.HasPassword || repo.clearedPasswordUserID != testTargetUserID {
		t.Fatalf("unexpected cleared user: %#v", user)
	}
	if !slices.Contains(repo.auditActions, auditActionClearPassword) {
		t.Fatalf("expected password removal audit event, got %#v", repo.auditActions)
	}
}

type fakeAdminRepository struct {
	admins                map[string]bool
	systemAdmins          []SystemAdmin
	managedUsers          []ManagedUser
	systemAdminByEmail    map[string]SystemAdmin
	systemAdminByID       map[string]ManagedUser
	revokeResult          SystemAdminRevokeResult
	activeAdminCount      int64
	grantedEmail          string
	grantedUserID         string
	revokedUserID         string
	disabledUserID        string
	disabledValue         bool
	passwordUserID        string
	clearedPasswordUserID string
	passwordHash          string
	auditActions          []string
	auditKeys             []string
}

func (r *fakeAdminRepository) IsSystemAdmin(_ context.Context, userID string) (bool, error) {
	return r.admins[userID], nil
}

func (r *fakeAdminRepository) ListSystemAdmins(context.Context) ([]SystemAdmin, error) {
	return append([]SystemAdmin(nil), r.systemAdmins...), nil
}

func (r *fakeAdminRepository) ListManagedUsers(context.Context) ([]ManagedUser, error) {
	return append([]ManagedUser(nil), r.managedUsers...), nil
}

func (r *fakeAdminRepository) GrantSystemAdminByEmail(_ context.Context, email string) (SystemAdmin, error) {
	r.grantedEmail = email
	admin, ok := r.systemAdminByEmail[email]
	if !ok {
		return SystemAdmin{}, errors.New("not found")
	}
	return admin, nil
}

func (r *fakeAdminRepository) GrantSystemAdminByUserID(_ context.Context, userID string) (ManagedUser, error) {
	r.grantedUserID = userID
	user, ok := r.systemAdminByID[userID]
	if !ok {
		return ManagedUser{}, errors.New("not found")
	}
	return user, nil
}

func (r *fakeAdminRepository) RevokeSystemAdminIfNotLast(_ context.Context, userID string) (SystemAdminRevokeResult, error) {
	r.revokedUserID = userID
	return r.revokeResult, nil
}

func (r *fakeAdminRepository) CountActiveSystemAdmins(context.Context) (int64, error) {
	if r.activeAdminCount == 0 {
		return int64(len(r.admins)), nil
	}
	return r.activeAdminCount, nil
}

func (r *fakeAdminRepository) SetManagedUserDisabledAt(_ context.Context, userID string, disabled bool) (ManagedUser, error) {
	r.disabledUserID = userID
	r.disabledValue = disabled
	for _, user := range r.managedUsers {
		if user.ID == userID {
			if disabled {
				now := time.Now().UTC()
				user.DisabledAt = &now
			} else {
				user.DisabledAt = nil
			}
			return user, nil
		}
	}
	return ManagedUser{}, errors.New("not found")
}

func (r *fakeAdminRepository) SetManagedUserPasswordHash(_ context.Context, userID, passwordHash string) (ManagedUser, error) {
	r.passwordUserID = userID
	r.passwordHash = passwordHash
	for _, user := range r.managedUsers {
		if user.ID == userID {
			return user, nil
		}
	}
	return ManagedUser{}, errors.New("not found")
}

func (r *fakeAdminRepository) ClearManagedUserPassword(_ context.Context, userID string) (ManagedUser, error) {
	r.clearedPasswordUserID = userID
	for _, user := range r.managedUsers {
		if user.ID == userID {
			user.HasPassword = false
			return user, nil
		}
	}
	return ManagedUser{}, errors.New("not found")
}

func (r *fakeAdminRepository) ExportData(context.Context) (DataExport, error) {
	return DataExport{Format: "netstamp.admin.data.v1"}, nil
}

func (r *fakeAdminRepository) ImportData(context.Context, DataExport) (DataImportResult, error) {
	return DataImportResult{}, nil
}

func (r *fakeAdminRepository) CreateSystemSettingAuditEvent(_ context.Context, key, action string, _ *string) error {
	r.auditKeys = append(r.auditKeys, key)
	r.auditActions = append(r.auditActions, action)
	return nil
}

type fakeAuthenticationMethodRepository struct {
	hasPassword   bool
	identityCount int64
}

func (r fakeAuthenticationMethodRepository) CountUserAuthenticationMethods(context.Context, string) (bool, int64, error) {
	return r.hasPassword, r.identityCount, nil
}
