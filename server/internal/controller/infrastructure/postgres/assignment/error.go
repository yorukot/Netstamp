package pgassignment

import (
	"fmt"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func mapAssignmentWriteError(err error) error {
	if postgres.IsForeignKeyViolation(err, "probe_check_assignments_project_id_fkey") {
		return fmt.Errorf("project not found: %w", domainproject.ErrProjectNotFound)
	}
	if postgres.IsForeignKeyViolation(err, "fk_probe_check_assignments_project_probe") {
		return fmt.Errorf("probe not found: %w", domainprobe.ErrProbeNotFound)
	}
	if postgres.IsForeignKeyViolation(err, "fk_probe_check_assignments_project_check") {
		return fmt.Errorf("check not found: %w", domaincheck.ErrCheckNotFound)
	}

	return err
}
