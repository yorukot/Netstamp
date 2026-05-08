package project

type Action string

const (
	ActionReadProject   Action = "read:project"
	ActionUpdateProject Action = "write:project"
	ActionDeleteProject Action = "delete:project"
	ActionManageMembers Action = "write:project_members"
	ActionManageLabels  Action = "write:project_labels"
	ActionManageChecks  Action = "write:project_checks"
	ActionCreateProbe   Action = "create:probe"
)

func Can(role Role, action Action) bool {
	switch action {
	case ActionReadProject:
		return IsValidRole(role)
	case ActionUpdateProject, ActionManageMembers:
		return role == RoleOwner || role == RoleAdmin
	case ActionManageLabels, ActionManageChecks, ActionCreateProbe:
		return role == RoleOwner || role == RoleAdmin || role == RoleEditor
	case ActionDeleteProject:
		return role == RoleOwner
	default:
		return false
	}
}

func CanAssignRole(actorRole, targetRole Role) bool {
	switch actorRole {
	case RoleOwner:
		return targetRole == RoleAdmin || targetRole == RoleEditor || targetRole == RoleViewer
	case RoleAdmin:
		return targetRole == RoleEditor || targetRole == RoleViewer
	default:
		return false
	}
}

func IsValidRole(role Role) bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleEditor, RoleViewer:
		return true
	default:
		return false
	}
}
