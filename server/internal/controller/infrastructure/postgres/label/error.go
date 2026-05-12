package pglabel

import (
	"fmt"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func mapLabelWriteError(err error) (bool, error) {
	if postgres.IsUniqueViolation(err, "uq_labels_active_project_key_value") {
		return true, fmt.Errorf("label already exists: %w", domainlabel.ErrLabelAlreadyExists)
	}
	if postgres.IsForeignKeyViolation(err, "labels_project_id_fkey") {
		return true, fmt.Errorf("project not found: %w", domainproject.ErrProjectNotFound)
	}

	return false, err
}
