package pglabel

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres/sqlc"
)

type LabelRepository struct {
	queries *sqlc.Queries
}

func NewLabelRepository(pool *pgxpool.Pool) *LabelRepository {
	return &LabelRepository{queries: sqlc.New(pool)}
}

func (r *LabelRepository) ListLabels(ctx context.Context, projectIDValue string) ([]domainlabel.Label, error) {
	ctx, span := postgres.StartDBSpan(ctx, pglabelTracer, "labels", "postgres.labels.list", "SELECT", "SELECT active labels for project")
	defer span.End()

	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListActiveLabelsForProject(ctx, projectID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	return mapLabels(rows), nil
}

func (r *LabelRepository) GetLabel(ctx context.Context, projectIDValue string, labelIDValue string) (domainlabel.Label, error) {
	ctx, span := postgres.StartDBSpan(ctx, pglabelTracer, "labels", "postgres.labels.select", "SELECT", "SELECT active label for project")
	defer span.End()

	projectID, labelID, err := parseProjectAndLabelIDs(projectIDValue, labelIDValue)
	if err != nil {
		return domainlabel.Label{}, err
	}

	row, err := r.queries.GetActiveLabelForProject(ctx, sqlc.GetActiveLabelForProjectParams{
		ProjectID: projectID,
		ID:        labelID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainlabel.Label{}, domainlabel.ErrLabelNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return domainlabel.Label{}, err
	}

	return mapLabel(row), nil
}

func (r *LabelRepository) CreateLabel(ctx context.Context, input domainlabel.CreateLabelStorageInput) (domainlabel.Label, error) {
	ctx, span := postgres.StartDBSpan(ctx, pglabelTracer, "labels", "postgres.labels.insert", "INSERT", "INSERT label")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domainlabel.Label{}, err
	}

	row, err := r.queries.CreateLabel(ctx, sqlc.CreateLabelParams{
		ProjectID: projectID,
		Key:       input.Key,
		Value:     input.Value,
	})
	if err != nil {
		mapped := mapLabelWriteError(err)
		if mapped == err {
			postgres.RecordDBSpanError(span, err)
		}
		return domainlabel.Label{}, mapped
	}

	return mapLabel(row), nil
}

func (r *LabelRepository) UpdateLabel(ctx context.Context, input domainlabel.UpdateLabelStorageInput) (domainlabel.Label, error) {
	ctx, span := postgres.StartDBSpan(ctx, pglabelTracer, "labels", "postgres.labels.update", "UPDATE", "UPDATE label")
	defer span.End()

	projectID, labelID, err := parseProjectAndLabelIDs(input.ProjectID, input.LabelID)
	if err != nil {
		return domainlabel.Label{}, err
	}

	row, err := r.queries.UpdateLabel(ctx, sqlc.UpdateLabelParams{
		ProjectID: projectID,
		ID:        labelID,
		Key:       input.Key,
		Value:     input.Value,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainlabel.Label{}, domainlabel.ErrLabelNotFound
		}
		mapped := mapLabelWriteError(err)
		if mapped == err {
			postgres.RecordDBSpanError(span, err)
		}
		return domainlabel.Label{}, mapped
	}

	return mapLabel(row), nil
}

func (r *LabelRepository) SoftDeleteLabel(ctx context.Context, projectIDValue string, labelIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pglabelTracer, "labels", "postgres.labels.soft_delete", "UPDATE", "SOFT DELETE label")
	defer span.End()

	projectID, labelID, err := parseProjectAndLabelIDs(projectIDValue, labelIDValue)
	if err != nil {
		return err
	}

	if _, err := r.queries.SoftDeleteLabel(ctx, sqlc.SoftDeleteLabelParams{
		ProjectID: projectID,
		ID:        labelID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainlabel.ErrLabelNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *LabelRepository) GetActiveLabelsByIDsForProject(ctx context.Context, projectIDValue string, labelIDValues []string) ([]domainlabel.Label, error) {
	ctx, span := postgres.StartDBSpan(ctx, pglabelTracer, "labels", "postgres.labels.select_by_ids", "SELECT", "SELECT active labels by ids for project")
	defer span.End()

	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return nil, err
	}
	labelIDs, err := ParseLabelIDs(labelIDValues)
	if err != nil {
		return nil, err
	}
	if len(labelIDs) == 0 {
		return []domainlabel.Label{}, nil
	}

	rows, err := r.queries.GetActiveLabelsByIDsForProject(ctx, sqlc.GetActiveLabelsByIDsForProjectParams{
		ProjectID: projectID,
		LabelIds:  labelIDs,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}
	if len(rows) != len(labelIDs) {
		return nil, domainlabel.ErrLabelNotFound
	}

	return mapLabels(rows), nil
}

func ParseLabelIDs(values []string) ([]uuid.UUID, error) {
	if len(values) == 0 {
		return nil, nil
	}

	labelIDs := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		labelID, err := postgres.ParseUUID(value, domainlabel.ErrInvalidInput)
		if err != nil {
			return nil, err
		}
		labelIDs = append(labelIDs, labelID)
	}

	return labelIDs, nil
}

func parseProjectAndLabelIDs(projectIDValue string, labelIDValue string) (uuid.UUID, uuid.UUID, error) {
	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	labelID, err := postgres.ParseUUID(labelIDValue, domainlabel.ErrLabelNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return projectID, labelID, nil
}
