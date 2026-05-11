package pgproject

import (
	"fmt"

	authapp "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	projectapp "github.com/yorukot/netstamp/internal/controller/application/project"
)

func mapCreateProjectError(err error) (bool, error) {
	if postgres.IsUniqueViolation(err, "uq_projects_slug") {
		return true, fmt.Errorf("project slug already exists: %w", projectapp.ErrProjectSlugAlreadyExists)
	}
	if postgres.IsForeignKeyViolation(err, "projects_created_by_user_id_fkey") {
		return true, fmt.Errorf("user not found: %w", authapp.ErrUserNotFound)
	}

	return false, err
}

func mapCreateProjectMemberError(err error) (bool, error) {
	if postgres.IsUniqueViolation(err, "uq_project_members_project_user") {
		return true, fmt.Errorf("project member already exists: %w", projectapp.ErrMemberAlreadyExists)
	}
	if postgres.IsForeignKeyViolation(err, "project_members_project_id_fkey") {
		return true, fmt.Errorf("project not found: %w", projectapp.ErrProjectNotFound)
	}
	if postgres.IsForeignKeyViolation(err, "project_members_user_id_fkey") {
		return true, fmt.Errorf("user not found: %w", authapp.ErrUserNotFound)
	}

	return false, err
}
