package pgpublicstatus

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

func mapNoRows(err, notFound error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return notFound
	}
	return err
}

func mapPublicStatusWriteError(err error) error {
	switch {
	case postgres.IsUniqueViolation(err, "uq_public_status_pages_slug"):
		return domainpublic.ErrSlugAlreadyExist
	case postgres.IsForeignKeyViolation(err, "public_status_pages_project_id_fkey"),
		postgres.IsForeignKeyViolation(err, "public_status_pages_created_by_user_id_fkey"),
		postgres.IsForeignKeyViolation(err, "fk_public_status_page_elements_page_project"):
		return fmt.Errorf("project not found: %w", domainproject.ErrProjectNotFound)
	case postgres.IsForeignKeyViolation(err, "fk_public_status_page_elements_parent"):
		return fmt.Errorf("parent element not found: %w", domainpublic.ErrElementNotFound)
	case postgres.IsForeignKeyViolation(err, "fk_public_status_page_elements_check"):
		return fmt.Errorf("check not found: %w", domaincheck.ErrCheckNotFound)
	default:
		return err
	}
}
