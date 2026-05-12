package pglabel

import (
	"fmt"

	labelapp "github.com/yorukot/netstamp/internal/controller/application/label"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
)

func mapLabelWriteError(err error) (bool, error) {
	if postgres.IsUniqueViolation(err, "uq_labels_active_project_key_value") {
		return true, fmt.Errorf("label already exists: %w", labelapp.ErrLabelAlreadyExists)
	}
	if postgres.IsForeignKeyViolation(err, "labels_project_id_fkey") {
		return true, fmt.Errorf("project not found: %w", labelapp.ErrProjectNotFound)
	}

	return false, err
}
