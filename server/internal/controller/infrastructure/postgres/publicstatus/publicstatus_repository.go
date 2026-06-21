package pgpublicstatus

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlc.New(pool)}
}

func (r *Repository) ListPages(ctx context.Context, projectIDValue string) ([]domainpublic.Page, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicstatusTracer, "public_status_pages", "postgres.public_status_pages.list", "SELECT", "SELECT public status pages")
	defer span.End()

	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListPublicStatusPages(ctx, projectID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}
	pages := make([]domainpublic.Page, 0, len(rows))
	for _, row := range rows {
		pages = append(pages, mapPage(row))
	}
	return pages, nil
}

func (r *Repository) GetPage(ctx context.Context, projectIDValue, pageIDValue string) (domainpublic.Page, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicstatusTracer, "public_status_pages", "postgres.public_status_pages.get", "SELECT", "SELECT public status page")
	defer span.End()

	projectID, pageID, err := parseProjectAndPageIDs(projectIDValue, pageIDValue)
	if err != nil {
		return domainpublic.Page{}, err
	}
	row, err := r.queries.GetPublicStatusPage(ctx, sqlc.GetPublicStatusPageParams{ProjectID: projectID, ID: pageID})
	if err != nil {
		return domainpublic.Page{}, mapNoRows(err, domainpublic.ErrPageNotFound)
	}
	return mapPage(row), nil
}

func (r *Repository) GetPageBySlug(ctx context.Context, slug string) (domainpublic.Page, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicstatusTracer, "public_status_pages", "postgres.public_status_pages.get_by_slug", "SELECT", "SELECT public status page by slug")
	defer span.End()

	row, err := r.queries.GetPublicStatusPageBySlug(ctx, slug)
	if err != nil {
		return domainpublic.Page{}, mapNoRows(err, domainpublic.ErrPageNotFound)
	}
	return mapPage(row), nil
}

func (r *Repository) CreatePage(ctx context.Context, input domainpublic.Page) (domainpublic.Page, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicstatusTracer, "public_status_pages", "postgres.public_status_pages.create", "INSERT", "INSERT public status page")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domainpublic.Page{}, err
	}
	createdByUserID, err := postgres.ParseUUID(input.CreatedByUserID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domainpublic.Page{}, err
	}
	row, err := r.queries.CreatePublicStatusPage(ctx, sqlc.CreatePublicStatusPageParams{
		ProjectID:         projectID,
		Slug:              input.Slug,
		Title:             input.Title,
		Description:       input.Description,
		Enabled:           input.Enabled,
		DefaultChartMode:  sqlcChartMode(input.DefaultChartMode),
		DefaultChartRange: sqlcChartRange(input.DefaultChartRange),
		CreatedByUserID:   createdByUserID,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainpublic.Page{}, mapPublicStatusWriteError(err)
	}
	return mapPage(row), nil
}

func (r *Repository) UpdatePage(ctx context.Context, input domainpublic.Page) (domainpublic.Page, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicstatusTracer, "public_status_pages", "postgres.public_status_pages.update", "UPDATE", "UPDATE public status page")
	defer span.End()

	projectID, pageID, err := parseProjectAndPageIDs(input.ProjectID, input.ID)
	if err != nil {
		return domainpublic.Page{}, err
	}
	row, err := r.queries.UpdatePublicStatusPage(ctx, sqlc.UpdatePublicStatusPageParams{
		ProjectID:         projectID,
		ID:                pageID,
		Slug:              input.Slug,
		Title:             input.Title,
		Description:       input.Description,
		Enabled:           input.Enabled,
		DefaultChartMode:  sqlcChartMode(input.DefaultChartMode),
		DefaultChartRange: sqlcChartRange(input.DefaultChartRange),
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainpublic.Page{}, mapPublicStatusWriteError(mapNoRows(err, domainpublic.ErrPageNotFound))
	}
	return mapPage(row), nil
}

func (r *Repository) DeletePage(ctx context.Context, projectIDValue, pageIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgpublicstatusTracer, "public_status_pages", "postgres.public_status_pages.soft_delete", "UPDATE", "SOFT DELETE public status page")
	defer span.End()

	projectID, pageID, err := parseProjectAndPageIDs(projectIDValue, pageIDValue)
	if err != nil {
		return err
	}
	rows, err := r.queries.SoftDeletePublicStatusPage(ctx, sqlc.SoftDeletePublicStatusPageParams{ProjectID: projectID, ID: pageID})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}
	if rows == 0 {
		return domainpublic.ErrPageNotFound
	}
	return nil
}

func (r *Repository) ListElements(ctx context.Context, pageIDValue string) ([]domainpublic.Element, error) {
	pageID, err := postgres.ParseUUID(pageIDValue, domainpublic.ErrPageNotFound)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListPublicStatusPageElements(ctx, pageID)
	if err != nil {
		return nil, err
	}
	elements := make([]domainpublic.Element, 0, len(rows))
	for _, row := range rows {
		elements = append(elements, mapListElement(row))
	}
	return elements, nil
}

func (r *Repository) GetElement(ctx context.Context, projectIDValue, pageIDValue, elementIDValue string) (domainpublic.Element, error) {
	projectID, pageID, elementID, err := parseElementScopeIDs(projectIDValue, pageIDValue, elementIDValue)
	if err != nil {
		return domainpublic.Element{}, err
	}
	row, err := r.queries.GetPublicStatusPageElement(ctx, sqlc.GetPublicStatusPageElementParams{
		ProjectID:    projectID,
		PublicPageID: pageID,
		ID:           elementID,
	})
	if err != nil {
		return domainpublic.Element{}, mapNoRows(err, domainpublic.ErrElementNotFound)
	}
	return mapElement(row), nil
}

func (r *Repository) CreateElement(ctx context.Context, input domainpublic.Element) (domainpublic.Element, error) {
	projectID, pageID, err := parseProjectAndPageIDs(input.ProjectID, input.PublicPageID)
	if err != nil {
		return domainpublic.Element{}, err
	}
	row, err := r.queries.CreatePublicStatusPageElement(ctx, sqlc.CreatePublicStatusPageElementParams{
		PublicPageID:    pageID,
		ProjectID:       projectID,
		ParentElementID: optionalUUID(input.ParentElementID, domainpublic.ErrElementNotFound),
		Kind:            sqlcElementKind(input.Kind),
		CheckID:         optionalUUID(input.CheckID, domainpublic.ErrInvalidInput),
		Title:           input.Title,
		Description:     input.Description,
		SortOrder:       input.SortOrder,
		ChartMode:       sqlcChartMode(input.ChartMode),
		ChartRange:      sqlcOptionalChartRange(input.ChartRange),
	})
	if err != nil {
		return domainpublic.Element{}, mapPublicStatusWriteError(err)
	}
	return mapElement(row), nil
}

func (r *Repository) UpdateElement(ctx context.Context, input domainpublic.Element) (domainpublic.Element, error) {
	projectID, pageID, elementID, err := parseElementScopeIDs(input.ProjectID, input.PublicPageID, input.ID)
	if err != nil {
		return domainpublic.Element{}, err
	}
	row, err := r.queries.UpdatePublicStatusPageElement(ctx, sqlc.UpdatePublicStatusPageElementParams{
		PublicPageID:    pageID,
		ProjectID:       projectID,
		ID:              elementID,
		ParentElementID: optionalUUID(input.ParentElementID, domainpublic.ErrElementNotFound),
		Kind:            sqlcElementKind(input.Kind),
		CheckID:         optionalUUID(input.CheckID, domainpublic.ErrInvalidInput),
		Title:           input.Title,
		Description:     input.Description,
		SortOrder:       input.SortOrder,
		ChartMode:       sqlcChartMode(input.ChartMode),
		ChartRange:      sqlcOptionalChartRange(input.ChartRange),
	})
	if err != nil {
		return domainpublic.Element{}, mapPublicStatusWriteError(mapNoRows(err, domainpublic.ErrElementNotFound))
	}
	return mapElement(row), nil
}

func (r *Repository) DeleteElement(ctx context.Context, projectIDValue, pageIDValue, elementIDValue string) error {
	projectID, pageID, elementID, err := parseElementScopeIDs(projectIDValue, pageIDValue, elementIDValue)
	if err != nil {
		return err
	}
	rows, err := r.queries.DeletePublicStatusPageElement(ctx, sqlc.DeletePublicStatusPageElementParams{ProjectID: projectID, PublicPageID: pageID, ID: elementID})
	if err != nil {
		return err
	}
	if rows == 0 {
		return domainpublic.ErrElementNotFound
	}
	return nil
}

func (r *Repository) ListAssignments(ctx context.Context, pageIDValue string) ([]domainpublic.Assignment, error) {
	pageID, err := postgres.ParseUUID(pageIDValue, domainpublic.ErrPageNotFound)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListPublicStatusAssignments(ctx, pageID)
	if err != nil {
		return nil, err
	}
	assignments := make([]domainpublic.Assignment, 0, len(rows))
	for _, row := range rows {
		assignments = append(assignments, mapAssignment(row))
	}
	return assignments, nil
}

func (r *Repository) ListIncidents(ctx context.Context, pageIDValue string, limit int32) ([]domainpublic.Incident, error) {
	pageID, err := postgres.ParseUUID(pageIDValue, domainpublic.ErrPageNotFound)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.queries.ListPublicStatusIncidents(ctx, sqlc.ListPublicStatusIncidentsParams{PublicPageID: pageID, LimitCount: limit})
	if err != nil {
		return nil, err
	}
	incidents := make([]domainpublic.Incident, 0, len(rows))
	for _, row := range rows {
		incidents = append(incidents, mapIncident(row))
	}
	return incidents, nil
}

func parseProjectAndPageIDs(projectIDValue, pageIDValue string) (uuid.UUID, uuid.UUID, error) {
	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	pageID, err := postgres.ParseUUID(pageIDValue, domainpublic.ErrPageNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return projectID, pageID, nil
}

func parseElementScopeIDs(projectIDValue, pageIDValue, elementIDValue string) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	projectID, pageID, err := parseProjectAndPageIDs(projectIDValue, pageIDValue)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	elementID, err := postgres.ParseUUID(elementIDValue, domainpublic.ErrElementNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	return projectID, pageID, elementID, nil
}

func optionalUUID(value *string, invalidErr error) *uuid.UUID {
	if value == nil {
		return nil
	}
	id, err := postgres.ParseUUID(*value, invalidErr)
	if err != nil {
		return nil
	}
	return &id
}
