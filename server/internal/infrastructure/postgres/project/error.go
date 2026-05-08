package pgproject

import (
	"fmt"

	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres"
)

func mapCreateProjectError(err error) error {
	if postgres.IsUniqueViolation(err, "uq_projects_slug") {
		return fmt.Errorf("project slug already exists: %w", domainproject.ErrProjectSlugAlreadyExists)
	}
	if postgres.IsForeignKeyViolation(err, "projects_created_by_user_id_fkey") {
		return fmt.Errorf("user not found: %w", identity.ErrUserNotFound)
	}

	return err
}

func mapCreateProjectMemberError(err error) error {
	if postgres.IsUniqueViolation(err, "uq_project_members_active_project_user") {
		return fmt.Errorf("project member already exists: %w", domainproject.ErrMemberAlreadyExists)
	}
	if postgres.IsForeignKeyViolation(err, "project_members_project_id_fkey") {
		return fmt.Errorf("project not found: %w", domainproject.ErrProjectNotFound)
	}
	if postgres.IsForeignKeyViolation(err, "project_members_user_id_fkey") {
		return fmt.Errorf("user not found: %w", identity.ErrUserNotFound)
	}

	return err
}
