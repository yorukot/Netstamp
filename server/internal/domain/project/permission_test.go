package project

import "testing"

func TestCan(t *testing.T) {
	tests := []struct {
		name   string
		role   Role
		action Action
		want   bool
	}{
		{name: "owner can read project", role: RoleOwner, action: ActionReadProject, want: true},
		{name: "admin can read project", role: RoleAdmin, action: ActionReadProject, want: true},
		{name: "editor can read project", role: RoleEditor, action: ActionReadProject, want: true},
		{name: "viewer can read project", role: RoleViewer, action: ActionReadProject, want: true},
		{name: "unknown cannot read project", role: Role("unknown"), action: ActionReadProject, want: false},
		{name: "owner can update project", role: RoleOwner, action: ActionUpdateProject, want: true},
		{name: "admin can update project", role: RoleAdmin, action: ActionUpdateProject, want: true},
		{name: "editor cannot update project", role: RoleEditor, action: ActionUpdateProject, want: false},
		{name: "viewer cannot update project", role: RoleViewer, action: ActionUpdateProject, want: false},
		{name: "owner can delete project", role: RoleOwner, action: ActionDeleteProject, want: true},
		{name: "admin cannot delete project", role: RoleAdmin, action: ActionDeleteProject, want: false},
		{name: "owner can manage members", role: RoleOwner, action: ActionManageMembers, want: true},
		{name: "admin can manage members", role: RoleAdmin, action: ActionManageMembers, want: true},
		{name: "editor cannot manage members", role: RoleEditor, action: ActionManageMembers, want: false},
		{name: "owner can manage labels", role: RoleOwner, action: ActionManageLabels, want: true},
		{name: "admin can manage labels", role: RoleAdmin, action: ActionManageLabels, want: true},
		{name: "editor can manage labels", role: RoleEditor, action: ActionManageLabels, want: true},
		{name: "viewer cannot manage labels", role: RoleViewer, action: ActionManageLabels, want: false},
		{name: "owner can manage checks", role: RoleOwner, action: ActionManageChecks, want: true},
		{name: "admin can manage checks", role: RoleAdmin, action: ActionManageChecks, want: true},
		{name: "editor can manage checks", role: RoleEditor, action: ActionManageChecks, want: true},
		{name: "viewer cannot manage checks", role: RoleViewer, action: ActionManageChecks, want: false},
		{name: "owner can manage probes", role: RoleOwner, action: ActionManageProbes, want: true},
		{name: "admin can manage probes", role: RoleAdmin, action: ActionManageProbes, want: true},
		{name: "editor can manage probes", role: RoleEditor, action: ActionManageProbes, want: true},
		{name: "viewer cannot manage probes", role: RoleViewer, action: ActionManageProbes, want: false},
		{name: "owner can create probe", role: RoleOwner, action: ActionCreateProbe, want: true},
		{name: "admin can create probe", role: RoleAdmin, action: ActionCreateProbe, want: true},
		{name: "editor can create probe", role: RoleEditor, action: ActionCreateProbe, want: true},
		{name: "viewer cannot create probe", role: RoleViewer, action: ActionCreateProbe, want: false},
		{name: "unknown action denied", role: RoleOwner, action: Action("unknown"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Can(tt.role, tt.action); got != tt.want {
				t.Fatalf("expected %t, got %t", tt.want, got)
			}
		})
	}
}

func TestActionValuesUseVerbResourceFormat(t *testing.T) {
	tests := map[Action]string{
		ActionReadProject:   "read:project",
		ActionUpdateProject: "write:project",
		ActionDeleteProject: "delete:project",
		ActionManageMembers: "write:project_members",
		ActionManageLabels:  "write:project_labels",
		ActionManageChecks:  "write:project_checks",
		ActionManageProbes:  "write:project_probes",
		ActionCreateProbe:   "create:probe",
	}

	for action, want := range tests {
		if got := string(action); got != want {
			t.Fatalf("expected action value %q, got %q", want, got)
		}
	}
}

func TestCanAssignRole(t *testing.T) {
	tests := []struct {
		name       string
		actorRole  Role
		targetRole Role
		want       bool
	}{
		{name: "owner cannot assign owner", actorRole: RoleOwner, targetRole: RoleOwner, want: false},
		{name: "owner can assign admin", actorRole: RoleOwner, targetRole: RoleAdmin, want: true},
		{name: "owner can assign editor", actorRole: RoleOwner, targetRole: RoleEditor, want: true},
		{name: "owner can assign viewer", actorRole: RoleOwner, targetRole: RoleViewer, want: true},
		{name: "admin cannot assign owner", actorRole: RoleAdmin, targetRole: RoleOwner, want: false},
		{name: "admin cannot assign admin", actorRole: RoleAdmin, targetRole: RoleAdmin, want: false},
		{name: "admin can assign editor", actorRole: RoleAdmin, targetRole: RoleEditor, want: true},
		{name: "admin can assign viewer", actorRole: RoleAdmin, targetRole: RoleViewer, want: true},
		{name: "editor cannot assign viewer", actorRole: RoleEditor, targetRole: RoleViewer, want: false},
		{name: "viewer cannot assign viewer", actorRole: RoleViewer, targetRole: RoleViewer, want: false},
		{name: "unknown actor cannot assign viewer", actorRole: Role("unknown"), targetRole: RoleViewer, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanAssignRole(tt.actorRole, tt.targetRole); got != tt.want {
				t.Fatalf("expected %t, got %t", tt.want, got)
			}
		})
	}
}

func TestIsValidRole(t *testing.T) {
	tests := []struct {
		role Role
		want bool
	}{
		{role: RoleOwner, want: true},
		{role: RoleAdmin, want: true},
		{role: RoleEditor, want: true},
		{role: RoleViewer, want: true},
		{role: Role("unknown"), want: false},
		{role: "", want: false},
	}

	for _, tt := range tests {
		if got := IsValidRole(tt.role); got != tt.want {
			t.Fatalf("expected role %q validity %t, got %t", tt.role, tt.want, got)
		}
	}
}
