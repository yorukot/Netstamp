package pgcheck

import (
	"fmt"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func mapCheckWriteError(err error) error {
	if postgres.IsForeignKeyViolation(err, "checks_project_id_fkey") {
		return fmt.Errorf("project not found: %w", domainproject.ErrProjectNotFound)
	}
	if postgres.IsForeignKeyViolation(err, "ping_check_configs_check_id_fkey") ||
		postgres.IsForeignKeyViolation(err, "traceroute_check_configs_check_id_fkey") {
		return fmt.Errorf("check not found: %w", domaincheck.ErrCheckNotFound)
	}
	if postgres.IsForeignKeyViolation(err, "fk_check_labels_project_check") {
		return fmt.Errorf("check not found: %w", domaincheck.ErrCheckNotFound)
	}
	if postgres.IsForeignKeyViolation(err, "fk_check_labels_project_label") {
		return fmt.Errorf("check label not found: %w", domainlabel.ErrLabelNotFound)
	}

	return err
}
