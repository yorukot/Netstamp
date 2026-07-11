package check

import (
	"context"
	"testing"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func TestCanManageChecksUsesProjectPermissionPolicy(t *testing.T) {
	tests := []struct {
		role domainproject.Role
		want bool
	}{
		{role: domainproject.RoleOwner, want: true},
		{role: domainproject.RoleAdmin, want: true},
		{role: domainproject.RoleEditor, want: true},
		{role: domainproject.RoleViewer, want: false},
	}

	for _, test := range tests {
		t.Run(string(test.role), func(t *testing.T) {
			service := &Service{projectAccess: staticRoleProjectAccess{role: test.role}}
			got, err := service.canManageChecks(context.Background(), "project-id", "user-id")
			if err != nil {
				t.Fatalf("resolve check capability: %v", err)
			}
			if got != test.want {
				t.Fatalf("expected capability %t, got %t", test.want, got)
			}
		})
	}
}

type staticRoleProjectAccess struct{ role domainproject.Role }

func (access staticRoleProjectAccess) GetProjectForUser(context.Context, string, string) (domainproject.Project, error) {
	return domainproject.Project{}, nil
}

func (access staticRoleProjectAccess) GetMemberRole(context.Context, string, string) (domainproject.Role, error) {
	return access.role, nil
}
