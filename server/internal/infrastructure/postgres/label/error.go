package pglabel

import (
	"fmt"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres"
)

func mapLabelWriteError(err error) error {
	if postgres.IsUniqueViolation(err, "uq_labels_active_project_key_value") {
		return fmt.Errorf("label already exists: %w", domainlabel.ErrLabelAlreadyExists)
	}
	if postgres.IsForeignKeyViolation(err, "labels_project_id_fkey") {
		return fmt.Errorf("project not found: %w", domainproject.ErrProjectNotFound)
	}

	return err
}
