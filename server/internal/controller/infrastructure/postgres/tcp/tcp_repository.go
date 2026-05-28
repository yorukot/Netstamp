package pgtcp

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

func (r *TCPRepository) CreateTCPResults(ctx context.Context, inputs []domaintcp.ResultStorageInput) error {
	ctx, span := postgres.StartDBSpan(ctx, pgtcpTracer, "tcp_results", "postgres.tcp_results.create_batch", "INSERT", "INSERT tcp result batch")
	defer span.End()

	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)
		for _, input := range inputs {
			createErr := q.CreateTCPResult(ctx, sqlc.CreateTCPResultParams{
				ProbeStorageID:    input.ProbeStorageID,
				CheckStorageID:    input.CheckStorageID,
				StartedAt:         timestamptz(input.StartedAt),
				FinishedAt:        timestamptz(input.FinishedAt),
				DurationMs:        input.DurationMs,
				Status:            sqlcTCPStatus(input.Status),
				ConnectDurationMs: input.ConnectDurationMs,
				ResolvedIp:        input.ResolvedIP,
				IpFamily:          sqlcIPFamily(input.IPFamily),
				ErrorCode:         input.ErrorCode,
				ErrorMessage:      input.ErrorMessage,
			})
			if createErr != nil {
				return mapTCPResultWriteError(createErr)
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

func (r *TCPRepository) ListTCPInsight(ctx context.Context, input domaintcp.InsightQuery) (domaintcp.InsightResult, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgtcpTracer, "tcp_results", "postgres.tcp_results.insight", "SELECT", "SELECT tcp insight")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domaintcp.InsightResult{}, err
	}
	probeID, err := postgres.ParseUUID(input.ProbeID, domainprobe.ErrInvalidInput)
	if err != nil {
		return domaintcp.InsightResult{}, err
	}
	checkID, err := postgres.ParseUUID(input.CheckID, domaincheck.ErrInvalidInput)
	if err != nil {
		return domaintcp.InsightResult{}, err
	}
	storageIDs, err := r.queries.ResolveTCPInsightStorageIDs(ctx, sqlc.ResolveTCPInsightStorageIDsParams{
		CheckID:   checkID,
		ProjectID: projectID,
		ProbeID:   probeID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domaintcp.InsightResult{}, domainprobe.ErrProbeNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return domaintcp.InsightResult{}, err
	}

	startedAtFrom := pgtype.Timestamptz{Time: input.From.UTC(), Valid: true}
	startedAtTo := pgtype.Timestamptz{Time: input.To.UTC(), Valid: true}
	countParams := sqlc.CountTCPInsightPointsParams{
		ProbeStorageID: storageIDs.ProbeStorageID,
		CheckStorageID: storageIDs.CheckStorageID,
		StartedAtFrom:  startedAtFrom,
		StartedAtTo:    startedAtTo,
	}
	totalPoints, err := r.queries.CountTCPInsightPoints(ctx, countParams)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domaintcp.InsightResult{}, err
	}

	summary, err := r.tcpInsightSummary(ctx, storageIDs.ProbeStorageID, storageIDs.CheckStorageID, startedAtFrom, startedAtTo)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domaintcp.InsightResult{}, err
	}

	if totalPoints <= int64(input.MaxDataPoints) {
		buckets, rawErr := r.listRawTCPInsight(ctx, storageIDs.ProbeStorageID, storageIDs.CheckStorageID, startedAtFrom, startedAtTo)
		if rawErr != nil {
			postgres.RecordDBSpanError(span, rawErr)
			return domaintcp.InsightResult{}, rawErr
		}
		return domaintcp.InsightResult{
			Buckets:     buckets,
			Summary:     summary,
			Resolution:  domaintcp.InsightResolutionRaw,
			TotalPoints: totalPoints,
		}, nil
	}

	buckets, bucketErr := r.listBucketTCPInsight(ctx, storageIDs.ProbeStorageID, storageIDs.CheckStorageID, startedAtFrom, startedAtTo, input.MaxDataPoints)
	if bucketErr != nil {
		postgres.RecordDBSpanError(span, bucketErr)
		return domaintcp.InsightResult{}, bucketErr
	}
	return domaintcp.InsightResult{
		Buckets:     buckets,
		Summary:     summary,
		Resolution:  domaintcp.InsightResolutionBucket,
		TotalPoints: totalPoints,
	}, nil
}

func (r *TCPRepository) tcpInsightSummary(ctx context.Context, probeStorageID, checkStorageID int64, startedAtFrom, startedAtTo pgtype.Timestamptz) (domaintcp.InsightSummary, error) {
	row, err := r.queries.GetTCPInsightSummary(ctx, sqlc.GetTCPInsightSummaryParams{
		ProbeStorageID: probeStorageID,
		CheckStorageID: checkStorageID,
		StartedAtFrom:  startedAtFrom,
		StartedAtTo:    startedAtTo,
	})
	if err != nil {
		return domaintcp.InsightSummary{}, err
	}
	return tcpInsightSummary(row), nil
}

func (r *TCPRepository) listRawTCPInsight(ctx context.Context, probeStorageID, checkStorageID int64, startedAtFrom, startedAtTo pgtype.Timestamptz) ([]domaintcp.InsightBucket, error) {
	rows, err := r.queries.ListTCPInsightRawRows(ctx, sqlc.ListTCPInsightRawRowsParams{
		ProbeStorageID: probeStorageID,
		CheckStorageID: checkStorageID,
		StartedAtFrom:  startedAtFrom,
		StartedAtTo:    startedAtTo,
	})
	if err != nil {
		return nil, err
	}
	return rawInsightRows(rows), nil
}

func (r *TCPRepository) listBucketTCPInsight(ctx context.Context, probeStorageID, checkStorageID int64, startedAtFrom, startedAtTo pgtype.Timestamptz, maxDataPoints int32) ([]domaintcp.InsightBucket, error) {
	rows, err := r.queries.ListTCPInsightBucketRows(ctx, sqlc.ListTCPInsightBucketRowsParams{
		StartedAtTo:    startedAtTo,
		StartedAtFrom:  startedAtFrom,
		MaxDataPoints:  float64(maxDataPoints),
		ProbeStorageID: probeStorageID,
		CheckStorageID: checkStorageID,
	})
	if err != nil {
		return nil, err
	}
	return bucketInsightRows(rows), nil
}

func rawInsightRows(rows []sqlc.ListTCPInsightRawRowsRow) []domaintcp.InsightBucket {
	buckets := make([]domaintcp.InsightBucket, 0, len(rows))
	for _, row := range rows {
		buckets = append(buckets, domaintcp.InsightBucket{
			Timestamp:       time.UnixMilli(row.BucketMs).UTC(),
			ResultCount:     row.ResultCount,
			DurationAvgMs:   floatPtr(row.DurationAvgMs),
			ConnectMinMs:    row.ConnectMinMs,
			ConnectAvgMs:    row.ConnectAvgMs,
			ConnectMedianMs: row.ConnectMedianMs,
			ConnectMaxMs:    row.ConnectMaxMs,
			ConnectStddevMs: row.ConnectStddevMs,
			SuccessRate:     floatPtr(row.SuccessRate),
			TimeoutCount:    row.TimeoutCount,
			ErrorCount:      row.ErrorCount,
		})
	}
	return buckets
}

func bucketInsightRows(rows []sqlc.ListTCPInsightBucketRowsRow) []domaintcp.InsightBucket {
	buckets := make([]domaintcp.InsightBucket, 0, len(rows))
	for _, row := range rows {
		buckets = append(buckets, domaintcp.InsightBucket{
			Timestamp:       time.UnixMilli(row.BucketMs).UTC(),
			ResultCount:     row.ResultCount,
			DurationAvgMs:   floatPtr(row.DurationAvgMs),
			ConnectMinMs:    floatPtrIf(row.ConnectValueCount, row.ConnectMinMs),
			ConnectAvgMs:    floatPtrIf(row.ConnectValueCount, row.ConnectAvgMs),
			ConnectMedianMs: floatPtrIf(row.ConnectValueCount, row.ConnectMedianMs),
			ConnectMaxMs:    floatPtrIf(row.ConnectValueCount, row.ConnectMaxMs),
			ConnectStddevMs: floatPtrIf(row.ConnectValueCount, row.ConnectStddevMs),
			SuccessRate:     floatPtr(row.SuccessRate),
			TimeoutCount:    row.TimeoutCount,
			ErrorCount:      row.ErrorCount,
		})
	}
	return buckets
}

func tcpInsightSummary(row sqlc.GetTCPInsightSummaryRow) domaintcp.InsightSummary {
	return domaintcp.InsightSummary{
		TotalResults:     row.TotalResults,
		SuccessfulCount:  row.SuccessfulCount,
		TimeoutCount:     row.TimeoutCount,
		ErrorCount:       row.ErrorCount,
		AvgConnectMs:     floatPtrIf(row.ConnectValueCount, row.AvgConnectMs),
		MedianConnectMs:  floatPtrIf(row.ConnectValueCount, row.MedianConnectMs),
		MaxConnectMs:     floatPtrIf(row.ConnectValueCount, row.MaxConnectMs),
		P95ConnectMs:     floatPtrIf(row.ConnectValueCount, row.P95ConnectMs),
		P99ConnectMs:     floatPtrIf(row.ConnectValueCount, row.P99ConnectMs),
		LatestStatus:     tcpStatusPtr(row.LatestStatus),
		LatestStartedAt:  timeMillisPtr(row.LatestStartedAtMs),
		LatestConnectMs:  row.LatestConnectMs,
		LatestResolvedIP: row.LatestResolvedIp,
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

func tcpStatusPtr(value string) *domaintcp.Status {
	if value == "" {
		return nil
	}
	status := domaintcp.Status(value)
	return &status
}

func timeMillisPtr(value int64) *time.Time {
	if value == 0 {
		return nil
	}
	timestamp := time.UnixMilli(value).UTC()
	return &timestamp
}
