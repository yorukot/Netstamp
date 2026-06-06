package pgping

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type PingRepository struct {
	queries *sqlc.Queries
	tx      *postgres.Transactor
}

func NewPingRepository(pool *pgxpool.Pool) *PingRepository {
	return &PingRepository{
		queries: sqlc.New(pool),
		tx:      postgres.NewTransactor(pool),
	}
}

func (r *PingRepository) CreatePingResults(ctx context.Context, inputs []domainping.ResultStorageInput) error {
	ctx, span := postgres.StartDBSpan(ctx, pgpingTracer, "ping_results", "postgres.ping_results.create_batch", "INSERT", "INSERT ping result batch")
	defer span.End()

	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)
		for _, input := range inputs {
			createErr := q.CreatePingResult(ctx, sqlc.CreatePingResultParams{
				ProbeStorageID: input.ProbeStorageID,
				CheckStorageID: input.CheckStorageID,
				StartedAt:      input.StartedAt.UTC(),
				FinishedAt:     input.FinishedAt.UTC(),
				DurationMs:     input.DurationMs,
				Status:         sqlcPingStatus(input.Status),
				SentCount:      input.SentCount,
				ReceivedCount:  input.ReceivedCount,
				LossPercent:    input.LossPercent,
				RttMinMs:       input.RttMinMs,
				RttAvgMs:       input.RttAvgMs,
				RttMedianMs:    input.RttMedianMs,
				RttMaxMs:       input.RttMaxMs,
				RttStddevMs:    input.RttStddevMs,
				RttSamplesMs:   storageRTTSamples(input.RttSamplesMs),
				ResolvedIp:     input.ResolvedIP,
				IpFamily:       sqlcIPFamily(input.IPFamily),
				ErrorCode:      input.ErrorCode,
				ErrorMessage:   input.ErrorMessage,
			})
			if createErr != nil {
				return mapPingResultWriteError(createErr)
			}
		}

		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func storageRTTSamples(samples []float64) []float64 {
	return append([]float64{}, samples...)
}

type pingSeriesScope struct {
	probeStorageID int64
	checkStorageID int64
	startedAtFrom  time.Time
	startedAtTo    time.Time
	maxDataPoints  int32
}

type pingSeriesQuery func(context.Context, postgres.SeriesScope) ([]domainping.SeriesPoint, error)

type pingSeriesQueries struct {
	raw    pingSeriesQuery
	bucket pingSeriesQuery
	rollup pingSeriesQuery
}

func (r *PingRepository) CountPingSeriesPoints(ctx context.Context, input domainping.SeriesPointCountQuery) (int64, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpingTracer, "ping_results", "postgres.ping_results.series_count", "SELECT", "SELECT ping result point count")
	defer span.End()

	probeStorageID, checkStorageID, err := r.resolvePingStorageIDs(ctx, input.ProjectID, input.ProbeID, input.CheckID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return 0, err
	}

	rawPoints, err := r.queries.CountPingResultSeriesPoints(ctx, sqlc.CountPingResultSeriesPointsParams{
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

func (r *PingRepository) ListPingSeries(ctx context.Context, input domainping.SeriesReadQuery) (map[string]domainping.SeriesData, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpingTracer, "ping_results", "postgres.ping_results.series", "SELECT", "SELECT ping result time series")
	defer span.End()

	probeStorageID, checkStorageID, err := r.resolvePingStorageIDs(ctx, input.ProjectID, input.ProbeID, input.CheckID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}
	scope := pingSeriesScope{
		probeStorageID: probeStorageID,
		checkStorageID: checkStorageID,
		startedAtFrom:  input.From.UTC(),
		startedAtTo:    input.To.UTC(),
		maxDataPoints:  input.MaxDataPoints,
	}

	series := make(map[string]domainping.SeriesData, len(input.Series))
	for _, key := range input.Series {
		points, listErr := r.listPingSeriesByKey(ctx, key, input.Mode, scope)
		if listErr != nil {
			postgres.RecordDBSpanError(span, listErr)
			return nil, listErr
		}
		series[key] = domainping.SeriesData{Points: points}
	}

	return series, nil
}

func (r *PingRepository) GetPingInsightSummary(ctx context.Context, input domainping.InsightSummaryQuery) (domainping.InsightSummary, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpingTracer, "ping_results", "postgres.ping_results.insight_summary", "SELECT", "SELECT ping insight summary")
	defer span.End()

	probeStorageID, checkStorageID, err := r.resolvePingStorageIDs(ctx, input.ProjectID, input.ProbeID, input.CheckID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainping.InsightSummary{}, err
	}
	scope := pingSeriesScope{
		probeStorageID: probeStorageID,
		checkStorageID: checkStorageID,
		startedAtFrom:  input.From.UTC(),
		startedAtTo:    input.To.UTC(),
	}

	switch input.Source {
	case domainping.SeriesSourceRaw:
		summary, summaryErr := r.pingInsightSummary(ctx, scope)
		if summaryErr != nil {
			postgres.RecordDBSpanError(span, summaryErr)
			return domainping.InsightSummary{}, summaryErr
		}
		return summary, nil
	case domainping.SeriesSourceAggregate:
		summary, summaryErr := r.pingRollupInsightSummary(ctx, scope)
		if summaryErr != nil {
			postgres.RecordDBSpanError(span, summaryErr)
			return domainping.InsightSummary{}, summaryErr
		}
		return summary, nil
	default:
		err := errors.New("unsupported ping insight summary source")
		postgres.RecordDBSpanError(span, err)
		return domainping.InsightSummary{}, err
	}
}

func (r *PingRepository) resolvePingStorageIDs(ctx context.Context, projectIDValue, probeIDValue, checkIDValue string) (int64, int64, error) {
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
	storageIDs, err := r.queries.ResolvePingSeriesStorageIDs(ctx, sqlc.ResolvePingSeriesStorageIDsParams{
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

func (r *PingRepository) listPingSeriesByKey(ctx context.Context, key string, mode domainping.SeriesReadMode, scope pingSeriesScope) ([]domainping.SeriesPoint, error) {
	queries, err := r.pingSeriesQueries(key)
	if err != nil {
		return nil, err
	}
	return listPingMetricSeries(ctx, mode, scope, queries)
}

func (r *PingRepository) pingSeriesQueries(key string) (pingSeriesQueries, error) {
	switch key {
	case "latency_avg":
		return pingSeriesQueries{
			raw:    postgres.NewRawSeriesQuery(r.queries.ListPingLatencyAvgRawSeries, latencyAvgRawSeriesRows),
			bucket: postgres.NewBucketSeriesQuery(r.queries.ListPingLatencyAvgBucketSeries, latencyAvgBucketSeriesRows),
			rollup: postgres.NewRollupSeriesQuery(r.queries.ListPingLatencyAvgRollupSeries, latencyAvgRollupSeriesRows),
		}, nil
	case "latency_min":
		return pingSeriesQueries{
			raw:    postgres.NewRawSeriesQuery(r.queries.ListPingLatencyMinRawSeries, latencyMinRawSeriesRows),
			bucket: postgres.NewBucketSeriesQuery(r.queries.ListPingLatencyMinBucketSeries, latencyMinBucketSeriesRows),
			rollup: postgres.NewRollupSeriesQuery(r.queries.ListPingLatencyMinRollupSeries, latencyMinRollupSeriesRows),
		}, nil
	case "latency_max":
		return pingSeriesQueries{
			raw:    postgres.NewRawSeriesQuery(r.queries.ListPingLatencyMaxRawSeries, latencyMaxRawSeriesRows),
			bucket: postgres.NewBucketSeriesQuery(r.queries.ListPingLatencyMaxBucketSeries, latencyMaxBucketSeriesRows),
			rollup: postgres.NewRollupSeriesQuery(r.queries.ListPingLatencyMaxRollupSeries, latencyMaxRollupSeriesRows),
		}, nil
	case "loss_percent":
		return pingSeriesQueries{
			raw:    postgres.NewRawSeriesQuery(r.queries.ListPingLossPercentRawSeries, lossPercentRawSeriesRows),
			bucket: postgres.NewBucketSeriesQuery(r.queries.ListPingLossPercentBucketSeries, lossPercentBucketSeriesRows),
			rollup: postgres.NewRollupSeriesQuery(r.queries.ListPingLossPercentRollupSeries, lossPercentRollupSeriesRows),
		}, nil
	default:
		return pingSeriesQueries{}, errors.New("unsupported ping series")
	}
}

func listPingMetricSeries(ctx context.Context, mode domainping.SeriesReadMode, scope pingSeriesScope, queries pingSeriesQueries) ([]domainping.SeriesPoint, error) {
	queryScope := postgres.SeriesScope{
		ProbeStorageID: scope.probeStorageID,
		CheckStorageID: scope.checkStorageID,
		StartedAtFrom:  scope.startedAtFrom,
		StartedAtTo:    scope.startedAtTo,
		MaxDataPoints:  scope.maxDataPoints,
	}
	switch mode {
	case domainping.SeriesReadModeRaw:
		return queries.raw(ctx, queryScope)
	case domainping.SeriesReadModeBucket:
		return queries.bucket(ctx, queryScope)
	case domainping.SeriesReadModeRollup:
		return queries.rollup(ctx, queryScope)
	default:
		return nil, errors.New("unsupported ping series read mode")
	}
}

func (r *PingRepository) pingInsightSummary(ctx context.Context, scope pingSeriesScope) (domainping.InsightSummary, error) {
	row, err := r.queries.GetPingInsightSummary(ctx, sqlc.GetPingInsightSummaryParams{
		ProbeStorageID: scope.probeStorageID,
		CheckStorageID: scope.checkStorageID,
		StartedAtFrom:  scope.startedAtFrom,
		StartedAtTo:    scope.startedAtTo,
	})
	if err != nil {
		return domainping.InsightSummary{}, err
	}
	return pingInsightSummary(row), nil
}

func (r *PingRepository) pingRollupInsightSummary(ctx context.Context, scope pingSeriesScope) (domainping.InsightSummary, error) {
	row, err := r.queries.GetPingInsightRollupSummary(ctx, sqlc.GetPingInsightRollupSummaryParams{
		ProbeStorageID: scope.probeStorageID,
		CheckStorageID: scope.checkStorageID,
		StartedAtFrom:  scope.startedAtFrom,
		StartedAtTo:    scope.startedAtTo,
	})
	if err != nil {
		return domainping.InsightSummary{}, err
	}
	return pingRollupInsightSummary(row), nil
}
