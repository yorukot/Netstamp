package pgtcp

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

type TCPRepository struct {
	queries *sqlc.Queries
	tx      *postgres.Transactor
}

func NewTCPRepository(pool *pgxpool.Pool) *TCPRepository {
	return &TCPRepository{
		queries: sqlc.New(pool),
		tx:      postgres.NewTransactor(pool),
	}
}

func (r *TCPRepository) CreateTCPResults(ctx context.Context, inputs []domaintcp.ResultStorageInput) ([]domaintcp.ResultStorageInput, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgtcpTracer, "tcp_results", "postgres.tcp_results.create_batch", "INSERT", "INSERT tcp result batch")
	defer span.End()

	inserted := make([]domaintcp.ResultStorageInput, 0, len(inputs))
	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)
		for _, input := range inputs {
			_, createErr := q.CreateTCPResult(ctx, sqlc.CreateTCPResultParams{
				ProbeStorageID:    input.ProbeStorageID,
				CheckStorageID:    input.CheckStorageID,
				StartedAt:         input.StartedAt.UTC(),
				FinishedAt:        input.FinishedAt.UTC(),
				DurationMs:        input.DurationMs,
				Status:            sqlcTCPStatus(input.Status),
				ConnectDurationMs: input.ConnectDurationMs,
				ResolvedIp:        input.ResolvedIP,
				IpFamily:          sqlcIPFamily(input.IPFamily),
				ErrorCode:         input.ErrorCode,
				ErrorMessage:      input.ErrorMessage,
			})
			if errors.Is(createErr, pgx.ErrNoRows) {
				continue
			}
			if createErr != nil {
				return mapTCPResultWriteError(createErr)
			}
			inserted = append(inserted, input)
		}

		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	return inserted, nil
}

type tcpSeriesScope struct {
	probeStorageID int64
	checkStorageID int64
	startedAtFrom  time.Time
	startedAtTo    time.Time
	maxDataPoints  int32
}

type tcpSeriesQuery func(context.Context, postgres.SeriesScope) ([]domaintcp.SeriesPoint, error)

type tcpSeriesQueries struct {
	raw    tcpSeriesQuery
	bucket tcpSeriesQuery
	rollup tcpSeriesQuery
}

func (r *TCPRepository) CountTCPSeriesPoints(ctx context.Context, input domaintcp.SeriesPointCountQuery) (int64, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgtcpTracer, "tcp_results", "postgres.tcp_results.series_count", "SELECT", "SELECT tcp result point count")
	defer span.End()

	probeStorageID, checkStorageID, err := r.resolveTCPStorageIDs(ctx, input.ProjectID, input.ProbeID, input.CheckID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return 0, err
	}

	rawPoints, err := r.queries.CountTCPResultSeriesPoints(ctx, sqlc.CountTCPResultSeriesPointsParams{
		ProbeStorageID: probeStorageID,
		CheckStorageID: checkStorageID,
		StartedAtFrom:  input.From.UTC(),
		StartedAtTo:    input.To.UTC(),
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return 0, err
	}

	return rawPoints, nil
}

func (r *TCPRepository) ListTCPSeries(ctx context.Context, input domaintcp.SeriesReadQuery) (map[string]domaintcp.SeriesData, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgtcpTracer, "tcp_results", "postgres.tcp_results.series", "SELECT", "SELECT tcp result time series")
	defer span.End()

	probeStorageID, checkStorageID, err := r.resolveTCPStorageIDs(ctx, input.ProjectID, input.ProbeID, input.CheckID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}
	scope := tcpSeriesScope{
		probeStorageID: probeStorageID,
		checkStorageID: checkStorageID,
		startedAtFrom:  input.From.UTC(),
		startedAtTo:    input.To.UTC(),
		maxDataPoints:  input.MaxDataPoints,
	}

	series := make(map[string]domaintcp.SeriesData, len(input.Series))
	for _, key := range input.Series {
		points, listErr := r.listTCPSeriesByKey(ctx, key, input.Mode, scope)
		if listErr != nil {
			postgres.RecordDBSpanError(span, listErr)
			return nil, listErr
		}
		series[key] = domaintcp.SeriesData{Points: points}
	}

	return series, nil
}

func (r *TCPRepository) GetTCPInsightSummary(ctx context.Context, input domaintcp.InsightSummaryQuery) (domaintcp.InsightSummary, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgtcpTracer, "tcp_results", "postgres.tcp_results.insight_summary", "SELECT", "SELECT tcp insight summary")
	defer span.End()

	probeStorageID, checkStorageID, err := r.resolveTCPStorageIDs(ctx, input.ProjectID, input.ProbeID, input.CheckID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domaintcp.InsightSummary{}, err
	}
	scope := tcpSeriesScope{
		probeStorageID: probeStorageID,
		checkStorageID: checkStorageID,
		startedAtFrom:  input.From.UTC(),
		startedAtTo:    input.To.UTC(),
	}

	switch input.Source {
	case domaintcp.SeriesSourceRaw:
		summary, summaryErr := r.tcpInsightSummary(ctx, scope)
		if summaryErr != nil {
			postgres.RecordDBSpanError(span, summaryErr)
			return domaintcp.InsightSummary{}, summaryErr
		}
		return summary, nil
	case domaintcp.SeriesSourceAggregate:
		summary, summaryErr := r.tcpInsightRollupSummary(ctx, scope)
		if summaryErr != nil {
			postgres.RecordDBSpanError(span, summaryErr)
			return domaintcp.InsightSummary{}, summaryErr
		}
		return summary, nil
	default:
		err := errors.New("unsupported tcp insight summary source")
		postgres.RecordDBSpanError(span, err)
		return domaintcp.InsightSummary{}, err
	}
}

func (r *TCPRepository) resolveTCPStorageIDs(ctx context.Context, projectIDValue, probeIDValue, checkIDValue string) (int64, int64, error) {
	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return 0, 0, err
	}
	probeID, err := postgres.ParseUUID(probeIDValue, domainprobe.ErrInvalidInput)
	if err != nil {
		return 0, 0, err
	}
	checkID, err := postgres.ParseUUID(checkIDValue, domaincheck.ErrInvalidInput)
	if err != nil {
		return 0, 0, err
	}
	storageIDs, err := r.queries.ResolveTCPInsightStorageIDs(ctx, sqlc.ResolveTCPInsightStorageIDsParams{
		CheckID:   checkID,
		ProjectID: projectID,
		ProbeID:   probeID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, domainprobe.ErrProbeNotFound
		}
		return 0, 0, err
	}

	return storageIDs.ProbeStorageID, storageIDs.CheckStorageID, nil
}

func (r *TCPRepository) tcpInsightSummary(ctx context.Context, scope tcpSeriesScope) (domaintcp.InsightSummary, error) {
	row, err := r.queries.GetTCPInsightSummary(ctx, sqlc.GetTCPInsightSummaryParams{
		ProbeStorageID: scope.probeStorageID,
		CheckStorageID: scope.checkStorageID,
		StartedAtFrom:  scope.startedAtFrom,
		StartedAtTo:    scope.startedAtTo,
	})
	if err != nil {
		return domaintcp.InsightSummary{}, err
	}
	return tcpInsightSummary(row), nil
}

func (r *TCPRepository) tcpInsightRollupSummary(ctx context.Context, scope tcpSeriesScope) (domaintcp.InsightSummary, error) {
	row, err := r.queries.GetTCPInsightRollupSummary(ctx, sqlc.GetTCPInsightRollupSummaryParams{
		ProbeStorageID: scope.probeStorageID,
		CheckStorageID: scope.checkStorageID,
		StartedAtFrom:  scope.startedAtFrom,
		StartedAtTo:    scope.startedAtTo,
	})
	if err != nil {
		return domaintcp.InsightSummary{}, err
	}
	return tcpInsightRollupSummary(row), nil
}

func (r *TCPRepository) listTCPSeriesByKey(ctx context.Context, key string, mode domaintcp.SeriesReadMode, scope tcpSeriesScope) ([]domaintcp.SeriesPoint, error) {
	queries, err := r.tcpSeriesQueries(key)
	if err != nil {
		return nil, err
	}
	return listTCPMetricSeries(ctx, mode, scope, queries)
}

func (r *TCPRepository) tcpSeriesQueries(key string) (tcpSeriesQueries, error) {
	switch key {
	case "connect_avg":
		return tcpSeriesQueries{
			raw:    postgres.NewRawSeriesQuery(r.queries.ListTCPConnectAvgRawSeries, connectAvgRawSeriesRows),
			bucket: postgres.NewBucketSeriesQuery(r.queries.ListTCPConnectAvgBucketSeries, connectAvgBucketSeriesRows),
			rollup: postgres.NewRollupSeriesQuery(r.queries.ListTCPConnectAvgRollupSeries, connectAvgRollupSeriesRows),
		}, nil
	case "connect_min":
		return tcpSeriesQueries{
			raw:    postgres.NewRawSeriesQuery(r.queries.ListTCPConnectMinRawSeries, connectMinRawSeriesRows),
			bucket: postgres.NewBucketSeriesQuery(r.queries.ListTCPConnectMinBucketSeries, connectMinBucketSeriesRows),
			rollup: postgres.NewRollupSeriesQuery(r.queries.ListTCPConnectMinRollupSeries, connectMinRollupSeriesRows),
		}, nil
	case "connect_max":
		return tcpSeriesQueries{
			raw:    postgres.NewRawSeriesQuery(r.queries.ListTCPConnectMaxRawSeries, connectMaxRawSeriesRows),
			bucket: postgres.NewBucketSeriesQuery(r.queries.ListTCPConnectMaxBucketSeries, connectMaxBucketSeriesRows),
			rollup: postgres.NewRollupSeriesQuery(r.queries.ListTCPConnectMaxRollupSeries, connectMaxRollupSeriesRows),
		}, nil
	case "failure_percent":
		return tcpSeriesQueries{
			raw:    postgres.NewRawSeriesQuery(r.queries.ListTCPFailurePercentRawSeries, failurePercentRawSeriesRows),
			bucket: postgres.NewBucketSeriesQuery(r.queries.ListTCPFailurePercentBucketSeries, failurePercentBucketSeriesRows),
			rollup: postgres.NewRollupSeriesQuery(r.queries.ListTCPFailurePercentRollupSeries, failurePercentRollupSeriesRows),
		}, nil
	default:
		return tcpSeriesQueries{}, errors.New("unsupported tcp series")
	}
}

func listTCPMetricSeries(ctx context.Context, mode domaintcp.SeriesReadMode, scope tcpSeriesScope, queries tcpSeriesQueries) ([]domaintcp.SeriesPoint, error) {
	queryScope := postgres.SeriesScope{
		ProbeStorageID: scope.probeStorageID,
		CheckStorageID: scope.checkStorageID,
		StartedAtFrom:  scope.startedAtFrom,
		StartedAtTo:    scope.startedAtTo,
		MaxDataPoints:  scope.maxDataPoints,
	}
	switch mode {
	case domaintcp.SeriesReadModeRaw:
		return queries.raw(ctx, queryScope)
	case domaintcp.SeriesReadModeBucket:
		return queries.bucket(ctx, queryScope)
	case domaintcp.SeriesReadModeRollup:
		return queries.rollup(ctx, queryScope)
	default:
		return nil, errors.New("unsupported tcp series read mode")
	}
}
