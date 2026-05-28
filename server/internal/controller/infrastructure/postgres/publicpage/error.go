package pgpublicpage

import (
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainpublicpage "github.com/yorukot/netstamp/internal/domain/publicpage"
)

func mapPublicPageWriteError(err error) error {
	switch {
	case postgres.IsUniqueViolation(err, "uq_public_pages_slug"):
		return domainpublicpage.ErrDuplicateSlug
	case postgres.IsUniqueViolation(err, "pk_public_page_folder_checks"):
		return domainpublicpage.ErrCheckAlreadyPublished
	case postgres.IsForeignKeyViolation(err, "public_pages_project_id_fkey"):
		return domainpublicpage.ErrPageNotFound
	case postgres.IsForeignKeyViolation(err, "fk_public_page_folders_parent"):
		return domainpublicpage.ErrFolderNotFound
	case postgres.IsForeignKeyViolation(err, "fk_public_page_folder_checks_folder"):
		return domainpublicpage.ErrFolderNotFound
	case postgres.IsForeignKeyViolation(err, "fk_public_page_folder_checks_check"):
		return domaincheck.ErrCheckNotFound
	default:
		return err
	}
}
