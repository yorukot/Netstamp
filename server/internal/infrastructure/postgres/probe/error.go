package pgprobe

import (
	"fmt"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres"
)

func mapCreateProbeError(err error) error {
	if postgres.IsForeignKeyViolation(err, "probes_project_id_fkey") {
		return fmt.Errorf("project not found: %w", domainproject.ErrProjectNotFound)
	}

	return err
}

func mapCreateProbeCredentialError(err error) error {
	if postgres.IsForeignKeyViolation(err, "probe_credentials_probe_id_fkey") {
		return fmt.Errorf("project not found: %w", domainproject.ErrProjectNotFound)
	}

	return err
}

func mapCreateProbeStatusError(err error) error {
	if postgres.IsForeignKeyViolation(err, "probe_statuses_probe_id_fkey") {
		return fmt.Errorf("project not found: %w", domainproject.ErrProjectNotFound)
	}

	return err
}

func mapCreateProbeLabelError(err error) error {
	if postgres.IsForeignKeyViolation(err, "fk_probe_labels_project_probe") {
		return fmt.Errorf("project not found: %w", domainproject.ErrProjectNotFound)
	}
	if postgres.IsForeignKeyViolation(err, "fk_probe_labels_project_label") {
		return fmt.Errorf("probe label not found: %w", domainprobe.ErrLabelNotFound)
	}

	return err
}
