package pgresult

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainresult "github.com/yorukot/netstamp/internal/domain/result"
)

type ResultRepository struct {
	queries *sqlc.Queries
}

func NewResultRepository(pool *pgxpool.Pool) *ResultRepository {
	return &ResultRepository{queries: sqlc.New(pool)}
}

func (r *ResultRepository) ListLatestResults(ctx context.Context, input domainresult.LatestResultQuery) (domainresult.LatestResultList, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgresultTracer, "results", "postgres.results.latest", "SELECT", "SELECT latest project results")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domainresult.LatestResultList{}, err
	}
	probeID, err := optionalUUID(input.ProbeID, domainprobe.ErrInvalidInput)
	if err != nil {
		return domainresult.LatestResultList{}, err
	}
	checkID, err := optionalUUID(input.CheckID, domaincheck.ErrInvalidInput)
	if err != nil {
		return domainresult.LatestResultList{}, err
	}

	rows, err := r.queries.ListLatestResults(ctx, sqlc.ListLatestResultsParams{
		ProjectID:  projectID,
		ProbeID:    probeID,
		CheckID:    checkID,
		ResultType: latestResultTypeParam(input.Type),
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainresult.LatestResultList{}, err
	}

	return mapLatestResults(rows), nil
}

func optionalUUID(value string, invalidErr error) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil //nolint:nilnil // Nil means no optional UUID filter.
	}
	id, err := postgres.ParseUUID(value, invalidErr)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func latestResultTypeParam(value *domainresult.LatestResultType) *string {
	if value == nil {
		return nil
	}
	output := string(*value)
	return &output
}

func mapLatestResults(rows []sqlc.ListLatestResultsRow) domainresult.LatestResultList {
	results := make([]domainresult.LatestResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, domainresult.LatestResult{
			Type:            domainresult.LatestResultType(row.ResultType),
			ProbeID:         row.ProbeID.String(),
			CheckID:         row.CheckID.String(),
			LatestStartedAt: row.LatestStartedAt,
			LatestStatus:    row.LatestStatus,
		})
	}

	return domainresult.LatestResultList{Results: results}
}
