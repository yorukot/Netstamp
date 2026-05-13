package pgping

import (
	"context"
	"time"

	"github.com/google/uuid"
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
			projectID, err := postgres.ParseUUID(input.ProjectID, domainping.ErrInvalidResult)
			if err != nil {
				return err
			}
			probeID, err := postgres.ParseUUID(input.ProbeID, domainping.ErrInvalidResult)
			if err != nil {
				return err
			}
			checkID, err := postgres.ParseUUID(input.CheckID, domainping.ErrInvalidResult)
			if err != nil {
				return err
			}

			createErr := q.CreatePingResult(ctx, sqlc.CreatePingResultParams{
				ProjectID:     projectID,
				CheckID:       checkID,
				ProbeID:       probeID,
				StartedAt:     pgtype.Timestamptz{Time: input.StartedAt.UTC(), Valid: true},
				FinishedAt:    pgtype.Timestamptz{Time: input.FinishedAt.UTC(), Valid: true},
				DurationMs:    input.DurationMs,
				Status:        sqlcPingStatus(input.Status),
				SentCount:     input.SentCount,
				ReceivedCount: input.ReceivedCount,
				LossPercent:   input.LossPercent,
				RttMinMs:      input.RttMinMs,
				RttAvgMs:      input.RttAvgMs,
				RttMedianMs:   input.RttMedianMs,
				RttMaxMs:      input.RttMaxMs,
				RttStddevMs:   input.RttStddevMs,
				RttSamplesMs:  storageRTTSamples(input.RttSamplesMs),
				ResolvedIp:    input.ResolvedIP,
				IpFamily:      sqlcIPFamily(input.IPFamily),
				Raw:           input.Raw,
				ErrorCode:     input.ErrorCode,
				ErrorMessage:  input.ErrorMessage,
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

	startedAtFrom := pgtype.Timestamptz{Time: input.From.UTC(), Valid: true}
	startedAtTo := pgtype.Timestamptz{Time: input.To.UTC(), Valid: true}
	countParams := sqlc.CountPingResultSeriesPointsParams{
		ProjectID:     projectID,
		ProbeID:       probeID,
		CheckID:       checkID,
		StartedAtFrom: startedAtFrom,
		StartedAtTo:   startedAtTo,
		Metric:        input.Metric,
	}
	totalPoints, err := r.queries.CountPingResultSeriesPoints(ctx, countParams)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainping.SeriesResult{}, err
	}

	if totalPoints <= int64(input.MaxDataPoints) {
		points, rawErr := r.listRawPingSeries(ctx, projectID, probeID, checkID, startedAtFrom, startedAtTo, input.Metric)
		if rawErr != nil {
			postgres.RecordDBSpanError(span, rawErr)
			return domainping.SeriesResult{}, rawErr
		}
		return domainping.SeriesResult{Points: points, Resolution: domainping.SeriesResolutionRaw, TotalPoints: totalPoints}, nil
	}

	points, bucketErr := r.listBucketPingSeries(ctx, projectID, probeID, checkID, startedAtFrom, startedAtTo, input)
	if bucketErr != nil {
		postgres.RecordDBSpanError(span, bucketErr)
		return domainping.SeriesResult{}, bucketErr
	}
	return domainping.SeriesResult{Points: points, Resolution: domainping.SeriesResolutionBucket, TotalPoints: totalPoints}, nil
}

func (r *PingRepository) listRawPingSeries(ctx context.Context, projectID, probeID, checkID uuid.UUID, startedAtFrom, startedAtTo pgtype.Timestamptz, metric string) ([]domainping.SeriesPoint, error) {
	rows, err := r.queries.ListPingResultRawSeries(ctx, sqlc.ListPingResultRawSeriesParams{
		Metric:        metric,
		ProjectID:     projectID,
		ProbeID:       probeID,
		CheckID:       checkID,
		StartedAtFrom: startedAtFrom,
		StartedAtTo:   startedAtTo,
	})
	if err != nil {
		return nil, err
	}
	return rawSeriesRows(rows), nil
}

func (r *PingRepository) listBucketPingSeries(ctx context.Context, projectID, probeID, checkID uuid.UUID, startedAtFrom, startedAtTo pgtype.Timestamptz, input domainping.SeriesQuery) ([]domainping.SeriesPoint, error) {
	rows, err := r.queries.ListPingResultBucketSeries(ctx, sqlc.ListPingResultBucketSeriesParams{
		StartedAtTo:   startedAtTo,
		StartedAtFrom: startedAtFrom,
		MaxDataPoints: float64(input.MaxDataPoints),
		Metric:        input.Metric,
		ProjectID:     projectID,
		ProbeID:       probeID,
		CheckID:       checkID,
	})
	if err != nil {
		return nil, err
	}
	return bucketSeriesRows(rows), nil
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

func seriesPoint(timestampMs int64, value float64) domainping.SeriesPoint {
	return domainping.SeriesPoint{
		Timestamp: time.UnixMilli(timestampMs).UTC(),
		Value:     value,
	}
}
