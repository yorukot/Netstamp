package pghttpcheck

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

var tracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/infrastructure/postgres")

type Repository struct {
	queries *sqlc.Queries
	tx      *postgres.Transactor
}

func (r *Repository) CountHTTPSeriesPoints(ctx context.Context, input domainhttp.SeriesPointCountQuery) (int64, error) {
	ctx, span := postgres.StartDBSpan(ctx, tracer, "http_results", "postgres.http_results.series_count", "SELECT", "SELECT http result point count")
	defer span.End()

	probeID, checkID, err := r.resolveStorageIDs(ctx, input.ProjectID, input.ProbeID, input.CheckID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return 0, err
	}
	count, err := r.queries.CountHTTPResultSeriesPoints(ctx, sqlc.CountHTTPResultSeriesPointsParams{
		ProbeStorageID: probeID,
		CheckStorageID: checkID,
		StartedAtFrom:  input.From.UTC(),
		StartedAtTo:    input.To.UTC(),
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return 0, err
	}
	return count, nil
}

func (r *Repository) ListHTTPSeries(ctx context.Context, input domainhttp.SeriesReadQuery) (map[string]domainhttp.SeriesData, error) {
	ctx, span := postgres.StartDBSpan(ctx, tracer, "http_results", "postgres.http_results.series", "SELECT", "SELECT http result time series")
	defer span.End()

	probeID, checkID, err := r.resolveStorageIDs(ctx, input.ProjectID, input.ProbeID, input.CheckID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}
	result := make(map[string]domainhttp.SeriesData, len(input.Series))
	for _, metric := range input.Series {
		var points []domainhttp.SeriesPoint
		switch input.Mode {
		case domainhttp.SeriesReadModeRaw:
			rows, queryErr := r.queries.ListHTTPMetricRawSeries(ctx, sqlc.ListHTTPMetricRawSeriesParams{Metric: metric, ProbeStorageID: probeID, CheckStorageID: checkID, StartedAtFrom: input.From.UTC(), StartedAtTo: input.To.UTC()})
			if queryErr != nil {
				postgres.RecordDBSpanError(span, queryErr)
				return nil, queryErr
			}
			for _, row := range rows {
				points = append(points, domainhttp.SeriesPoint{Timestamp: time.UnixMilli(row.BucketMs).UTC(), Value: row.Value})
			}
		case domainhttp.SeriesReadModeBucket:
			rows, queryErr := r.queries.ListHTTPMetricBucketSeries(ctx, sqlc.ListHTTPMetricBucketSeriesParams{StartedAtTo: input.To.UTC(), StartedAtFrom: input.From.UTC(), MaxDataPoints: float64(input.MaxDataPoints), Metric: metric, ProbeStorageID: probeID, CheckStorageID: checkID})
			if queryErr != nil {
				postgres.RecordDBSpanError(span, queryErr)
				return nil, queryErr
			}
			for _, row := range rows {
				points = append(points, domainhttp.SeriesPoint{Timestamp: time.UnixMilli(row.BucketMs).UTC(), Value: row.Value})
			}
		case domainhttp.SeriesReadModeRollup:
			rows, queryErr := r.queries.ListHTTPMetricRollupSeries(ctx, sqlc.ListHTTPMetricRollupSeriesParams{Metric: metric, StartedAtTo: input.To.UTC(), StartedAtFrom: input.From.UTC(), MaxDataPoints: float64(input.MaxDataPoints), ProbeStorageID: probeID, CheckStorageID: checkID})
			if queryErr != nil {
				postgres.RecordDBSpanError(span, queryErr)
				return nil, queryErr
			}
			for _, row := range rows {
				points = append(points, domainhttp.SeriesPoint{Timestamp: time.UnixMilli(row.BucketMs).UTC(), Value: row.Value})
			}
		default:
			err := errors.New("unsupported http series mode")
			postgres.RecordDBSpanError(span, err)
			return nil, err
		}
		result[metric] = domainhttp.SeriesData{Points: points}
	}
	return result, nil
}

func (r *Repository) GetHTTPInsightSummary(ctx context.Context, input domainhttp.InsightSummaryQuery) (domainhttp.InsightSummary, error) {
	ctx, span := postgres.StartDBSpan(ctx, tracer, "http_results", "postgres.http_results.insight_summary", "SELECT", "SELECT http insight summary")
	defer span.End()

	probeID, checkID, err := r.resolveStorageIDs(ctx, input.ProjectID, input.ProbeID, input.CheckID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainhttp.InsightSummary{}, err
	}
	switch input.Source {
	case domainhttp.SeriesSourceAggregate:
		row, err := r.queries.GetHTTPInsightRollupSummary(ctx, sqlc.GetHTTPInsightRollupSummaryParams{ProbeStorageID: probeID, CheckStorageID: checkID, StartedAtFrom: input.From.UTC(), StartedAtTo: input.To.UTC()})
		if err != nil {
			postgres.RecordDBSpanError(span, err)
			return domainhttp.InsightSummary{}, err
		}
		return mapInsight(row.TotalResults, row.AverageTotalMs, row.MaxTotalMs, row.AverageTtfbMs, row.MaxTtfbMs, row.FailurePercent, row.SuccessRate, row.CertificateDaysRemaining, row.TtfbCount, row.CertificateCount, row.Samples), nil
	case domainhttp.SeriesSourceRaw:
		row, err := r.queries.GetHTTPInsightSummary(ctx, sqlc.GetHTTPInsightSummaryParams{ProbeStorageID: probeID, CheckStorageID: checkID, StartedAtFrom: input.From.UTC(), StartedAtTo: input.To.UTC()})
		if err != nil {
			postgres.RecordDBSpanError(span, err)
			return domainhttp.InsightSummary{}, err
		}
		return mapInsight(row.TotalResults, row.AverageTotalMs, row.MaxTotalMs, row.AverageTtfbMs, row.MaxTtfbMs, row.FailurePercent, row.SuccessRate, row.CertificateDaysRemaining, row.TtfbCount, row.CertificateCount, row.Samples), nil
	default:
		err := errors.New("unsupported http insight summary source")
		postgres.RecordDBSpanError(span, err)
		return domainhttp.InsightSummary{}, err
	}
}

func (r *Repository) resolveStorageIDs(ctx context.Context, projectValue, probeValue, checkValue string) (int64, int64, error) {
	projectID, err := postgres.ParseUUID(projectValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return 0, 0, err
	}
	probeID, err := postgres.ParseUUID(probeValue, domainprobe.ErrInvalidInput)
	if err != nil {
		return 0, 0, err
	}
	checkID, err := postgres.ParseUUID(checkValue, domaincheck.ErrInvalidInput)
	if err != nil {
		return 0, 0, err
	}
	row, err := r.queries.ResolveHTTPInsightStorageIDs(ctx, sqlc.ResolveHTTPInsightStorageIDsParams{CheckID: checkID, ProjectID: projectID, ProbeID: probeID})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, domainprobe.ErrProbeNotFound
	}
	if err != nil {
		return 0, 0, err
	}
	return row.ProbeStorageID, row.CheckStorageID, nil
}

func mapInsight(total int64, avgTotal, maxTotal, avgTTFB, maxTTFB, failure, success, cert float64, ttfbCount, certCount, samples int64) domainhttp.InsightSummary {
	value := func(enabled bool, number float64) *float64 {
		if !enabled {
			return nil
		}
		return &number
	}
	return domainhttp.InsightSummary{TotalResults: total, AverageTotalMs: value(total > 0, avgTotal), MaxTotalMs: value(total > 0, maxTotal), AverageTTFBMs: value(ttfbCount > 0, avgTTFB), MaxTTFBMs: value(ttfbCount > 0, maxTTFB), FailurePercent: value(total > 0, failure), SuccessRate: value(total > 0, success), CertificateDaysRemaining: value(certCount > 0, cert), Samples: samples}
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlc.New(pool), tx: postgres.NewTransactor(pool)}
}

func (r *Repository) CreateHTTPResults(ctx context.Context, inputs []domainhttp.ResultStorageInput) ([]domainhttp.ResultStorageInput, error) {
	ctx, span := postgres.StartDBSpan(ctx, tracer, "http_results", "postgres.http_results.create_batch", "INSERT", "INSERT http result batch")
	defer span.End()
	inserted := make([]domainhttp.ResultStorageInput, 0, len(inputs))
	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)
		for _, input := range inputs {
			_, err := q.CreateHTTPResult(ctx, sqlc.CreateHTTPResultParams{
				ProbeStorageID: input.ProbeStorageID, CheckStorageID: input.CheckStorageID,
				StartedAt: input.StartedAt.UTC(), FinishedAt: input.FinishedAt.UTC(),
				DurationMs: input.DurationMs, Status: sqlc.HttpStatus(input.Status),
				DnsDurationMs: input.DNSDurationMs, ConnectDurationMs: input.ConnectDurationMs,
				TlsDurationMs: input.TLSDurationMs, TtfbDurationMs: input.TTFBDurationMs,
				ResolvedIp: input.ResolvedIP, IpFamily: sqlcIPFamily(input.IPFamily),
				StatusCode: input.StatusCode, FinalUrl: input.FinalURL,
				RedirectCount: input.RedirectCount, ResponseBytes: input.ResponseBytes,
				ResponseTruncated: input.ResponseTruncated, BodyMatched: input.BodyMatched,
				TlsVersion: input.TLSVersion, TlsCipherSuite: input.TLSCipherSuite,
				CertificateNotBefore: input.CertificateNotBefore,
				CertificateNotAfter:  input.CertificateNotAfter,
				ErrorCode:            input.ErrorCode, ErrorMessage: input.ErrorMessage,
			})
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			if err != nil {
				return err
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

func sqlcIPFamily(value *domainnetwork.IPFamily) *sqlc.IpFamily {
	if value == nil {
		return nil
	}
	family := sqlc.IpFamily(*value)
	return &family
}
