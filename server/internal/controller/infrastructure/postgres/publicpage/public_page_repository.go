package pgpublicpage

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublicpage "github.com/yorukot/netstamp/internal/domain/publicpage"
)

type PublicPageRepository struct {
	queries *sqlc.Queries
	tx      *postgres.Transactor
}

func NewPublicPageRepository(pool *pgxpool.Pool) *PublicPageRepository {
	return &PublicPageRepository{
		queries: sqlc.New(pool),
		tx:      postgres.NewTransactor(pool),
	}
}

func (r *PublicPageRepository) ListPages(ctx context.Context, projectIDValue string) ([]domainpublicpage.Page, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicpageTracer, "public_pages", "postgres.public_pages.list", "SELECT", "SELECT public pages for project")
	defer span.End()

	projectID, err := parseProjectID(projectIDValue)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListPublicPagesForProject(ctx, projectID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	return mapPages(rows), nil
}

func (r *PublicPageRepository) GetPageForProject(ctx context.Context, projectIDValue, pageIDValue string) (domainpublicpage.Page, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicpageTracer, "public_pages", "postgres.public_pages.select", "SELECT", "SELECT active public page for project")
	defer span.End()

	projectID, pageID, err := parseProjectAndPageIDs(projectIDValue, pageIDValue)
	if err != nil {
		return domainpublicpage.Page{}, err
	}
	row, err := r.queries.GetActivePublicPageForProject(ctx, sqlc.GetActivePublicPageForProjectParams{
		ProjectID: projectID,
		ID:        pageID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainpublicpage.Page{}, domainpublicpage.ErrPageNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return domainpublicpage.Page{}, err
	}

	return mapPage(row), nil
}

func (r *PublicPageRepository) GetEnabledPageBySlug(ctx context.Context, slug string) (domainpublicpage.Page, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicpageTracer, "public_pages", "postgres.public_pages.public_select", "SELECT", "SELECT enabled public page by slug")
	defer span.End()

	row, err := r.queries.GetEnabledPublicPageBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainpublicpage.Page{}, domainpublicpage.ErrPageNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return domainpublicpage.Page{}, err
	}

	return mapPage(row), nil
}

func (r *PublicPageRepository) CreatePage(ctx context.Context, input domainpublicpage.Page) (domainpublicpage.Page, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicpageTracer, "public_pages", "postgres.public_pages.create", "INSERT", "INSERT public page")
	defer span.End()

	projectID, err := parseProjectID(input.ProjectID)
	if err != nil {
		return domainpublicpage.Page{}, err
	}
	row, err := r.queries.CreatePublicPage(ctx, sqlc.CreatePublicPageParams{
		ProjectID:   projectID,
		Slug:        input.Slug,
		Title:       input.Title,
		Description: input.Description,
		Enabled:     input.Enabled,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainpublicpage.Page{}, mapPublicPageWriteError(err)
	}

	return mapPage(row), nil
}

func (r *PublicPageRepository) UpdatePage(ctx context.Context, input domainpublicpage.PageUpdate) (domainpublicpage.Page, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicpageTracer, "public_pages", "postgres.public_pages.update", "UPDATE", "UPDATE public page")
	defer span.End()

	projectID, pageID, err := parseProjectAndPageIDs(input.ProjectID, input.ID)
	if err != nil {
		return domainpublicpage.Page{}, err
	}
	row, err := r.queries.UpdatePublicPage(ctx, sqlc.UpdatePublicPageParams{
		ProjectID:      projectID,
		ID:             pageID,
		Slug:           input.Slug,
		Title:          input.Title,
		DescriptionSet: input.DescriptionSet,
		Description:    input.Description,
		Enabled:        input.Enabled,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainpublicpage.Page{}, domainpublicpage.ErrPageNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return domainpublicpage.Page{}, mapPublicPageWriteError(err)
	}

	return mapPage(row), nil
}

func (r *PublicPageRepository) SoftDeletePage(ctx context.Context, projectIDValue, pageIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicpageTracer, "public_pages", "postgres.public_pages.soft_delete", "UPDATE", "SOFT DELETE public page")
	defer span.End()

	projectID, pageID, err := parseProjectAndPageIDs(projectIDValue, pageIDValue)
	if err != nil {
		return err
	}
	_, err = r.queries.SoftDeletePublicPage(ctx, sqlc.SoftDeletePublicPageParams{
		ProjectID: projectID,
		ID:        pageID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainpublicpage.ErrPageNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *PublicPageRepository) ListFolders(ctx context.Context, projectIDValue, pageIDValue string) ([]domainpublicpage.Folder, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicpageTracer, "public_page_folders", "postgres.public_page_folders.list", "SELECT", "SELECT public page folders")
	defer span.End()

	projectID, pageID, err := parseProjectAndPageIDs(projectIDValue, pageIDValue)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListPublicPageFoldersForProjectPage(ctx, sqlc.ListPublicPageFoldersForProjectPageParams{
		ProjectID:    projectID,
		PublicPageID: pageID,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	return mapFolders(rows), nil
}

func (r *PublicPageRepository) ListFolderChecks(ctx context.Context, projectIDValue, pageIDValue string) ([]domainpublicpage.PublishedCheck, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicpageTracer, "public_page_folder_checks", "postgres.public_page_folder_checks.list", "SELECT", "SELECT public page folder checks")
	defer span.End()

	projectID, pageID, err := parseProjectAndPageIDs(projectIDValue, pageIDValue)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListPublicPageFolderChecksForProjectPage(ctx, sqlc.ListPublicPageFolderChecksForProjectPageParams{
		ProjectID:    projectID,
		PublicPageID: pageID,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	return mapPublishedChecks(rows), nil
}

func (r *PublicPageRepository) ListPingPairs(ctx context.Context, projectIDValue, pageIDValue string) ([]domainpublicpage.PingPair, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicpageTracer, "public_page_folder_checks", "postgres.public_page_ping_pairs.list", "SELECT", "SELECT public ping pairs")
	defer span.End()

	projectID, pageID, err := parseProjectAndPageIDs(projectIDValue, pageIDValue)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListPublicPagePingPairs(ctx, sqlc.ListPublicPagePingPairsParams{
		ProjectID:    projectID,
		PublicPageID: pageID,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	return mapPingPairs(rows), nil
}

func (r *PublicPageRepository) CreateFolder(ctx context.Context, projectIDValue string, input domainpublicpage.Folder) (domainpublicpage.Folder, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicpageTracer, "public_page_folders", "postgres.public_page_folders.create", "INSERT", "INSERT public page folder")
	defer span.End()

	projectID, pageID, parentID, err := parseFolderWriteIDs(projectIDValue, input.PageID, input.ParentID)
	if err != nil {
		return domainpublicpage.Folder{}, err
	}
	row, err := r.queries.CreatePublicPageFolder(ctx, sqlc.CreatePublicPageFolderParams{
		ProjectID:    projectID,
		PublicPageID: pageID,
		ParentID:     parentID,
		Name:         input.Name,
		Description:  input.Description,
		SortOrder:    input.SortOrder,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainpublicpage.Folder{}, domainpublicpage.ErrPageNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return domainpublicpage.Folder{}, mapPublicPageWriteError(err)
	}

	return mapFolder(row), nil
}

func (r *PublicPageRepository) UpdateFolder(ctx context.Context, projectIDValue string, input domainpublicpage.FolderUpdate) (domainpublicpage.Folder, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicpageTracer, "public_page_folders", "postgres.public_page_folders.update", "UPDATE", "UPDATE public page folder")
	defer span.End()

	projectID, pageID, folderID, parentID, err := parseFolderUpdateIDs(projectIDValue, input.PageID, input.ID, input.ParentID)
	if err != nil {
		return domainpublicpage.Folder{}, err
	}
	row, err := r.queries.UpdatePublicPageFolder(ctx, sqlc.UpdatePublicPageFolderParams{
		ProjectID:      projectID,
		PublicPageID:   pageID,
		ID:             folderID,
		ParentIDSet:    input.ParentIDSet,
		ParentID:       parentID,
		Name:           input.Name,
		DescriptionSet: input.DescriptionSet,
		Description:    input.Description,
		SortOrder:      input.SortOrder,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainpublicpage.Folder{}, domainpublicpage.ErrFolderNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return domainpublicpage.Folder{}, mapPublicPageWriteError(err)
	}

	return mapFolder(row), nil
}

func (r *PublicPageRepository) DeleteFolder(ctx context.Context, projectIDValue, pageIDValue, folderIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicpageTracer, "public_page_folders", "postgres.public_page_folders.delete", "DELETE", "DELETE public page folder")
	defer span.End()

	projectID, pageID, folderID, err := parseProjectPageFolderIDs(projectIDValue, pageIDValue, folderIDValue)
	if err != nil {
		return err
	}
	_, err = r.queries.DeletePublicPageFolder(ctx, sqlc.DeletePublicPageFolderParams{
		ProjectID:    projectID,
		PublicPageID: pageID,
		ID:           folderID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainpublicpage.ErrFolderNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *PublicPageRepository) SetFolderChecks(ctx context.Context, projectIDValue, pageIDValue, folderIDValue string, checkIDValues []string) ([]domainpublicpage.PublishedCheck, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicpageTracer, "public_page_folder_checks", "postgres.public_page_folder_checks.replace", "DELETE INSERT", "REPLACE public page folder checks")
	defer span.End()

	projectID, pageID, folderID, err := parseProjectPageFolderIDs(projectIDValue, pageIDValue, folderIDValue)
	if err != nil {
		return nil, err
	}
	checkIDs, err := parseCheckIDs(checkIDValues)
	if err != nil {
		return nil, err
	}
	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)
		deleteErr := q.DeletePublicPageFolderChecks(ctx, sqlc.DeletePublicPageFolderChecksParams{
			PublicPageID: pageID,
			ProjectID:    projectID,
			FolderID:     folderID,
		})
		if deleteErr != nil {
			return deleteErr
		}
		for index, checkID := range checkIDs {
			_, createErr := q.CreatePublicPageFolderCheck(ctx, sqlc.CreatePublicPageFolderCheckParams{
				PublicPageID: pageID,
				ProjectID:    projectID,
				FolderID:     folderID,
				CheckID:      checkID,
				SortOrder:    int32(index),
			})
			if createErr != nil {
				if errors.Is(createErr, pgx.ErrNoRows) {
					return domainpublicpage.ErrInvalidInput
				}
				return mapPublicPageWriteError(createErr)
			}
		}
		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	checks, err := r.ListFolderChecks(ctx, projectIDValue, pageIDValue)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}
	filtered := checks[:0]
	for _, check := range checks {
		if check.FolderID == folderIDValue {
			filtered = append(filtered, check)
		}
	}
	return filtered, nil
}

func (r *PublicPageRepository) ResolvePublicPingPairProjectID(ctx context.Context, slug, probeIDValue, checkIDValue string) (string, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicpageTracer, "public_page_folder_checks", "postgres.public_page_ping_pair.resolve", "SELECT", "RESOLVE public ping pair")
	defer span.End()

	probeID, err := postgres.ParseUUID(probeIDValue, domainpublicpage.ErrInvalidInput)
	if err != nil {
		return "", err
	}
	checkID, err := postgres.ParseUUID(checkIDValue, domainpublicpage.ErrInvalidInput)
	if err != nil {
		return "", err
	}
	projectID, err := r.queries.ResolvePublicPingPairProjectID(ctx, sqlc.ResolvePublicPingPairProjectIDParams{
		Slug:    slug,
		ProbeID: probeID,
		CheckID: checkID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domainpublicpage.ErrCheckNotPublished
		}
		postgres.RecordDBSpanError(span, err)
		return "", err
	}

	return projectID.String(), nil
}

func parseProjectID(projectIDValue string) (uuid.UUID, error) {
	return postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
}

func parseProjectAndPageIDs(projectIDValue, pageIDValue string) (uuid.UUID, uuid.UUID, error) {
	projectID, err := parseProjectID(projectIDValue)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	pageID, err := postgres.ParseUUID(pageIDValue, domainpublicpage.ErrPageNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return projectID, pageID, nil
}

func parseProjectPageFolderIDs(projectIDValue, pageIDValue, folderIDValue string) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	projectID, pageID, err := parseProjectAndPageIDs(projectIDValue, pageIDValue)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	folderID, err := postgres.ParseUUID(folderIDValue, domainpublicpage.ErrFolderNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	return projectID, pageID, folderID, nil
}

func parseFolderWriteIDs(projectIDValue, pageIDValue string, parentIDValue *string) (uuid.UUID, uuid.UUID, *uuid.UUID, error) {
	projectID, pageID, err := parseProjectAndPageIDs(projectIDValue, pageIDValue)
	if err != nil {
		return uuid.Nil, uuid.Nil, nil, err
	}
	parentID, err := parseOptionalFolderID(parentIDValue)
	if err != nil {
		return uuid.Nil, uuid.Nil, nil, err
	}
	return projectID, pageID, parentID, nil
}

func parseFolderUpdateIDs(projectIDValue, pageIDValue, folderIDValue string, parentIDValue *string) (uuid.UUID, uuid.UUID, uuid.UUID, *uuid.UUID, error) {
	projectID, pageID, folderID, err := parseProjectPageFolderIDs(projectIDValue, pageIDValue, folderIDValue)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, nil, err
	}
	parentID, err := parseOptionalFolderID(parentIDValue)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, nil, err
	}
	return projectID, pageID, folderID, parentID, nil
}

func parseOptionalFolderID(folderIDValue *string) (*uuid.UUID, error) {
	if folderIDValue == nil {
		return nil, nil //nolint:nilnil // Nil means no parent folder was provided.
	}
	folderID, err := postgres.ParseUUID(*folderIDValue, domainpublicpage.ErrFolderNotFound)
	if err != nil {
		return nil, err
	}
	return &folderID, nil
}

func parseCheckIDs(checkIDValues []string) ([]uuid.UUID, error) {
	checkIDs := make([]uuid.UUID, 0, len(checkIDValues))
	for _, checkIDValue := range checkIDValues {
		checkID, err := postgres.ParseUUID(checkIDValue, domaincheck.ErrCheckNotFound)
		if err != nil {
			return nil, err
		}
		checkIDs = append(checkIDs, checkID)
	}
	return checkIDs, nil
}
