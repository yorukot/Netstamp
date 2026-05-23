package pgping

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
				StartedAt:      pgtype.Timestamptz{Time: input.StartedAt.UTC(), Valid: true},
				FinishedAt:     pgtype.Timestamptz{Time: input.FinishedAt.UTC(), Valid: true},
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

func (r *PingRepository) ListPingSeries(ctx context.Context, input domainping.SeriesQuery) (domainping.SeriesResult, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpingTracer, "ping_results", "postgres.ping_results.series", "SELECT", "SELECT ping result time series")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domainping.SeriesResult{}, err
	}
	probeID, err := postgres.ParseUUID(input.ProbeID, domainprobe.ErrInvalidInput)
	if err != nil {
		return domainping.SeriesResult{}, err
	}
	checkID, err := postgres.ParseUUID(input.CheckID, domaincheck.ErrInvalidInput)
	if err != nil {
		return domainping.SeriesResult{}, err
	}
	storageIDs, err := r.queries.ResolvePingSeriesStorageIDs(ctx, sqlc.ResolvePingSeriesStorageIDsParams{
		CheckID:   checkID,
		ProjectID: projectID,
		ProbeID:   probeID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainping.SeriesResult{}, domainprobe.ErrProbeNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return domainping.SeriesResult{}, err
	}

	startedAtFrom := pgtype.Timestamptz{Time: input.From.UTC(), Valid: true}
	startedAtTo := pgtype.Timestamptz{Time: input.To.UTC(), Valid: true}
	countParams := sqlc.CountPingResultSeriesPointsParams{
		ProbeStorageID: storageIDs.ProbeStorageID,
		CheckStorageID: storageIDs.CheckStorageID,
		StartedAtFrom:  startedAtFrom,
		StartedAtTo:    startedAtTo,
		Metric:         input.Metric,
	}
	totalPoints, err := r.queries.CountPingResultSeriesPoints(ctx, countParams)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainping.SeriesResult{}, err
	}

	if totalPoints <= int64(input.MaxDataPoints) {
		points, rawErr := r.listRawPingSeries(ctx, storageIDs.ProbeStorageID, storageIDs.CheckStorageID, startedAtFrom, startedAtTo, input.Metric)
		if rawErr != nil {
			postgres.RecordDBSpanError(span, rawErr)
			return domainping.SeriesResult{}, rawErr
		}
		return domainping.SeriesResult{Points: points, Resolution: domainping.SeriesResolutionRaw, TotalPoints: totalPoints}, nil
	}

	points, bucketErr := r.listBucketPingSeries(ctx, storageIDs.ProbeStorageID, storageIDs.CheckStorageID, startedAtFrom, startedAtTo, input)
	if bucketErr != nil {
		postgres.RecordDBSpanError(span, bucketErr)
		return domainping.SeriesResult{}, bucketErr
	}
	return domainping.SeriesResult{Points: points, Resolution: domainping.SeriesResolutionBucket, TotalPoints: totalPoints}, nil
}

func (r *PingRepository) ListPingInsight(ctx context.Context, input domainping.InsightQuery) (domainping.InsightResult, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgpingTracer, "ping_results", "postgres.ping_results.insight", "SELECT", "SELECT ping insight")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domainping.InsightResult{}, err
	}
	probeID, err := postgres.ParseUUID(input.ProbeID, domainprobe.ErrInvalidInput)
	if err != nil {
		return domainping.InsightResult{}, err
	}
	checkID, err := postgres.ParseUUID(input.CheckID, domaincheck.ErrInvalidInput)
	if err != nil {
		return domainping.InsightResult{}, err
	}
	storageIDs, err := r.queries.ResolvePingSeriesStorageIDs(ctx, sqlc.ResolvePingSeriesStorageIDsParams{
		CheckID:   checkID,
		ProjectID: projectID,
		ProbeID:   probeID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainping.InsightResult{}, domainprobe.ErrProbeNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return domainping.InsightResult{}, err
	}

	startedAtFrom := pgtype.Timestamptz{Time: input.From.UTC(), Valid: true}
	startedAtTo := pgtype.Timestamptz{Time: input.To.UTC(), Valid: true}
	countParams := sqlc.CountPingInsightPointsParams{
		ProbeStorageID: storageIDs.ProbeStorageID,
		CheckStorageID: storageIDs.CheckStorageID,
		StartedAtFrom:  startedAtFrom,
		StartedAtTo:    startedAtTo,
	}
	totalPoints, err := r.queries.CountPingInsightPoints(ctx, countParams)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainping.InsightResult{}, err
	}

	summary, err := r.pingInsightSummary(ctx, storageIDs.ProbeStorageID, storageIDs.CheckStorageID, startedAtFrom, startedAtTo)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainping.InsightResult{}, err
	}

	if totalPoints <= int64(input.MaxDataPoints) {
		buckets, cells, rawErr := r.listRawPingInsight(ctx, storageIDs.ProbeStorageID, storageIDs.CheckStorageID, startedAtFrom, startedAtTo)
		if rawErr != nil {
			postgres.RecordDBSpanError(span, rawErr)
			return domainping.InsightResult{}, rawErr
		}
		return domainping.InsightResult{
			Buckets:       buckets,
			SampleDensity: cells,
			Summary:       summary,
			Resolution:    domainping.SeriesResolutionRaw,
			TotalPoints:   totalPoints,
		}, nil
	}

	buckets, cells, bucketErr := r.listBucketPingInsight(ctx, storageIDs.ProbeStorageID, storageIDs.CheckStorageID, startedAtFrom, startedAtTo, input.MaxDataPoints)
	if bucketErr != nil {
		postgres.RecordDBSpanError(span, bucketErr)
		return domainping.InsightResult{}, bucketErr
	}
	return domainping.InsightResult{
		Buckets:       buckets,
		SampleDensity: cells,
		Summary:       summary,
		Resolution:    domainping.SeriesResolutionBucket,
		TotalPoints:   totalPoints,
	}, nil
}

func (r *PingRepository) listRawPingSeries(ctx context.Context, probeStorageID, checkStorageID int64, startedAtFrom, startedAtTo pgtype.Timestamptz, metric string) ([]domainping.SeriesPoint, error) {
	rows, err := r.queries.ListPingResultRawSeries(ctx, sqlc.ListPingResultRawSeriesParams{
		Metric:         metric,
		ProbeStorageID: probeStorageID,
		CheckStorageID: checkStorageID,
		StartedAtFrom:  startedAtFrom,
		StartedAtTo:    startedAtTo,
	})
	if err != nil {
		return nil, err
	}
	return rawSeriesRows(rows), nil
}

func (r *PingRepository) listBucketPingSeries(ctx context.Context, probeStorageID, checkStorageID int64, startedAtFrom, startedAtTo pgtype.Timestamptz, input domainping.SeriesQuery) ([]domainping.SeriesPoint, error) {
	rows, err := r.queries.ListPingResultBucketSeries(ctx, sqlc.ListPingResultBucketSeriesParams{
		StartedAtTo:    startedAtTo,
		StartedAtFrom:  startedAtFrom,
		MaxDataPoints:  float64(input.MaxDataPoints),
		Metric:         input.Metric,
		ProbeStorageID: probeStorageID,
		CheckStorageID: checkStorageID,
	})
	if err != nil {
		return nil, err
	}
	return bucketSeriesRows(rows), nil
}

func (r *PingRepository) pingInsightSummary(ctx context.Context, probeStorageID, checkStorageID int64, startedAtFrom, startedAtTo pgtype.Timestamptz) (domainping.InsightSummary, error) {
	row, err := r.queries.GetPingInsightSummary(ctx, sqlc.GetPingInsightSummaryParams{
		ProbeStorageID: probeStorageID,
		CheckStorageID: checkStorageID,
		StartedAtFrom:  startedAtFrom,
		StartedAtTo:    startedAtTo,
	})
	if err != nil {
		return domainping.InsightSummary{}, err
	}
	return pingInsightSummary(row), nil
}

func (r *PingRepository) listRawPingInsight(ctx context.Context, probeStorageID, checkStorageID int64, startedAtFrom, startedAtTo pgtype.Timestamptz) ([]domainping.InsightBucket, []domainping.SampleDensityCell, error) {
	rows, err := r.queries.ListPingInsightRawRows(ctx, sqlc.ListPingInsightRawRowsParams{
		ProbeStorageID: probeStorageID,
		CheckStorageID: checkStorageID,
		StartedAtFrom:  startedAtFrom,
		StartedAtTo:    startedAtTo,
	})
	if err != nil {
		return nil, nil, err
	}
	cells, err := r.queries.ListPingInsightRawSampleDensity(ctx, sqlc.ListPingInsightRawSampleDensityParams{
		ProbeStorageID: probeStorageID,
		CheckStorageID: checkStorageID,
		StartedAtFrom:  startedAtFrom,
		StartedAtTo:    startedAtTo,
	})
	if err != nil {
		return nil, nil, err
	}
	return rawInsightRows(rows), rawSampleDensityCells(cells), nil
}

func (r *PingRepository) listBucketPingInsight(ctx context.Context, probeStorageID, checkStorageID int64, startedAtFrom, startedAtTo pgtype.Timestamptz, maxDataPoints int32) ([]domainping.InsightBucket, []domainping.SampleDensityCell, error) {
	rows, err := r.queries.ListPingInsightBucketRows(ctx, sqlc.ListPingInsightBucketRowsParams{
		StartedAtTo:    startedAtTo,
		StartedAtFrom:  startedAtFrom,
		MaxDataPoints:  float64(maxDataPoints),
		ProbeStorageID: probeStorageID,
		CheckStorageID: checkStorageID,
	})
	if err != nil {
		return nil, nil, err
	}
	cells, err := r.queries.ListPingInsightBucketSampleDensity(ctx, sqlc.ListPingInsightBucketSampleDensityParams{
		StartedAtTo:    startedAtTo,
		StartedAtFrom:  startedAtFrom,
		MaxDataPoints:  float64(maxDataPoints),
		ProbeStorageID: probeStorageID,
		CheckStorageID: checkStorageID,
	})
	if err != nil {
		return nil, nil, err
	}
	return bucketInsightRows(rows), bucketSampleDensityCells(cells), nil
}

func rawSeriesRows(rows []sqlc.ListPingResultRawSeriesRow) []domainping.SeriesPoint {
	points := make([]domainping.SeriesPoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, seriesPoint(row.BucketMs, row.Value))
	}
	return points
}

func bucketSeriesRows(rows []sqlc.ListPingResultBucketSeriesRow) []domainping.SeriesPoint {
	points := make([]domainping.SeriesPoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, seriesPoint(row.BucketMs, row.Value))
	}
	return points
}

func rawInsightRows(rows []sqlc.ListPingInsightRawRowsRow) []domainping.InsightBucket {
	buckets := make([]domainping.InsightBucket, 0, len(rows))
	for _, row := range rows {
		buckets = append(buckets, domainping.InsightBucket{
			Timestamp:     time.UnixMilli(row.BucketMs).UTC(),
			ResultCount:   row.ResultCount,
			DurationAvgMs: floatPtr(row.DurationAvgMs),
			RttMinMs:      row.RttMinMs,
			RttAvgMs:      row.RttAvgMs,
			RttMedianMs:   row.RttMedianMs,
			RttMaxMs:      row.RttMaxMs,
			RttStddevMs:   row.RttStddevMs,
			LossPercent:   floatPtr(row.LossPercent),
			SuccessRate:   floatPtr(row.SuccessRate),
			SentCount:     row.SentCount,
			ReceivedCount: row.ReceivedCount,
			TimeoutCount:  row.TimeoutCount,
			ErrorCount:    row.ErrorCount,
		})
	}
	return buckets
}

func bucketInsightRows(rows []sqlc.ListPingInsightBucketRowsRow) []domainping.InsightBucket {
	buckets := make([]domainping.InsightBucket, 0, len(rows))
	for _, row := range rows {
		buckets = append(buckets, domainping.InsightBucket{
			Timestamp:     time.UnixMilli(row.BucketMs).UTC(),
			ResultCount:   row.ResultCount,
			DurationAvgMs: floatPtr(row.DurationAvgMs),
			RttMinMs:      floatPtrIf(row.RttValueCount, row.RttMinMs),
			RttAvgMs:      floatPtrIf(row.RttValueCount, row.RttAvgMs),
			RttMedianMs:   floatPtrIf(row.RttValueCount, row.RttMedianMs),
			RttMaxMs:      floatPtrIf(row.RttValueCount, row.RttMaxMs),
			RttStddevMs:   floatPtrIf(row.RttValueCount, row.RttStddevMs),
			LossPercent:   floatPtr(row.LossPercent),
			SuccessRate:   floatPtr(row.SuccessRate),
			SentCount:     row.SentCount,
			ReceivedCount: row.ReceivedCount,
			TimeoutCount:  row.TimeoutCount,
			ErrorCount:    row.ErrorCount,
		})
	}
	return buckets
}

func rawSampleDensityCells(rows []sqlc.ListPingInsightRawSampleDensityRow) []domainping.SampleDensityCell {
	cells := make([]domainping.SampleDensityCell, 0, len(rows))
	for _, row := range rows {
		cells = append(cells, sampleDensityCell(row.BucketMs, row.RttBucketStartMs, row.RttBucketEndMs, row.SampleCount))
	}
	return cells
}

func bucketSampleDensityCells(rows []sqlc.ListPingInsightBucketSampleDensityRow) []domainping.SampleDensityCell {
	cells := make([]domainping.SampleDensityCell, 0, len(rows))
	for _, row := range rows {
		cells = append(cells, sampleDensityCell(row.BucketMs, row.RttBucketStartMs, row.RttBucketEndMs, row.SampleCount))
	}
	return cells
}

func sampleDensityCell(bucketMs int64, rttBucketStartMs, rttBucketEndMs float64, sampleCount int64) domainping.SampleDensityCell {
	return domainping.SampleDensityCell{
		Timestamp:        time.UnixMilli(bucketMs).UTC(),
		RttBucketStartMs: rttBucketStartMs,
		RttBucketEndMs:   rttBucketEndMs,
		SampleCount:      sampleCount,
	}
}

func pingInsightSummary(row sqlc.GetPingInsightSummaryRow) domainping.InsightSummary {
	return domainping.InsightSummary{
		TotalResults:      row.TotalResults,
		SuccessfulCount:   row.SuccessfulCount,
		TimeoutCount:      row.TimeoutCount,
		ErrorCount:        row.ErrorCount,
		SentCount:         row.SentCount,
		ReceivedCount:     row.ReceivedCount,
		AvgLossPercent:    floatPtrIf(row.TotalResults, row.AvgLossPercent),
		AvgRttMs:          floatPtrIf(row.RttValueCount, row.AvgRttMs),
		MedianRttMs:       floatPtrIf(row.RttValueCount, row.MedianRttMs),
		MaxRttMs:          floatPtrIf(row.RttValueCount, row.MaxRttMs),
		P95RttMs:          floatPtrIf(row.SampleCount, row.P95RttMs),
		P99RttMs:          floatPtrIf(row.SampleCount, row.P99RttMs),
		LatestStatus:      pingStatusPtr(row.LatestStatus),
		LatestStartedAt:   timeMillisPtr(row.LatestStartedAtMs),
		LatestRttAvgMs:    row.LatestRttAvgMs,
		LatestLossPercent: floatPtrIf(row.TotalResults, row.LatestLossPercent),
		LatestResolvedIP:  row.LatestResolvedIp,
	}
}

func seriesPoint(timestampMs int64, value float64) domainping.SeriesPoint {
	return domainping.SeriesPoint{
		Timestamp: time.UnixMilli(timestampMs).UTC(),
		Value:     value,
	}
}

func floatPtr(value float64) *float64 {
	copied := value
	return &copied
}

func floatPtrIf(count int64, value float64) *float64 {
	if count == 0 {
		return nil
	}
	return floatPtr(value)
}

func pingStatusPtr(value string) *domainping.Status {
	if value == "" {
		return nil
	}
	status := domainping.Status(value)
	return &status
}

func timeMillisPtr(value int64) *time.Time {
	if value == 0 {
		return nil
	}
	timestamp := time.UnixMilli(value).UTC()
	return &timestamp
}
