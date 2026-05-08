package pgproject

import (
	"fmt"

	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres"
)

func mapCreateProjectError(err error) (bool, error) {
	if postgres.IsUniqueViolation(err, "uq_projects_slug") {
		return true, fmt.Errorf("project slug already exists: %w", domainproject.ErrProjectSlugAlreadyExists)
	}
	if postgres.IsForeignKeyViolation(err, "projects_created_by_user_id_fkey") {
		return true, fmt.Errorf("user not found: %w", identity.ErrUserNotFound)
	}

	return false, err
}

func mapCreateProjectMemberError(err error) (bool, error) {
	if postgres.IsUniqueViolation(err, "uq_project_members_active_project_user") {
		return true, fmt.Errorf("project member already exists: %w", domainproject.ErrMemberAlreadyExists)
	}
	if postgres.IsForeignKeyViolation(err, "project_members_project_id_fkey") {
		return true, fmt.Errorf("project not found: %w", domainproject.ErrProjectNotFound)
	}
	if postgres.IsForeignKeyViolation(err, "project_members_user_id_fkey") {
		return true, fmt.Errorf("user not found: %w", identity.ErrUserNotFound)
	}

	return false, err
}
