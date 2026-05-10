package pgprobe

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
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
		return fmt.Errorf("probe label not found: %w", domainlabel.ErrLabelNotFound)
	}

	return err
}

func mapUpdateProbeError(err error) error {
	if postgres.IsForeignKeyViolation(err, "fk_probe_labels_project_label") {
		return fmt.Errorf("probe label not found: %w", domainlabel.ErrLabelNotFound)
	}
	if postgres.IsForeignKeyViolation(err, "fk_effective_probe_checks_project_probe") {
		return fmt.Errorf("probe not found: %w", domainprobe.ErrProbeNotFound)
	}
	if postgres.IsForeignKeyViolation(err, "fk_effective_probe_checks_project_check") {
		return err
	}

	return err
}

func mapProbeLookupError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("probe not found: %w", domainprobe.ErrProbeNotFound)
	}

	return err
}
