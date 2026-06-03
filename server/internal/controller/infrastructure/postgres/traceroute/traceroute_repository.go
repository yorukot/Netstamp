package pgtraceroute

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type TracerouteRepository struct {
	queries *sqlc.Queries
	tx      *postgres.Transactor
}

func NewTracerouteRepository(pool *pgxpool.Pool) *TracerouteRepository {
	return &TracerouteRepository{
		queries: sqlc.New(pool),
		tx:      postgres.NewTransactor(pool),
	}
}

func (r *TracerouteRepository) CreateTracerouteResults(ctx context.Context, inputs []domaintraceroute.ResultStorageInput) error {
	ctx, span := postgres.StartDBSpan(ctx, pgtracerouteTracer, "traceroute_results", "postgres.traceroute_results.create_batch", "INSERT", "INSERT traceroute result batch")
	defer span.End()

	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)
		for _, input := range inputs {
			startedAt := input.StartedAt.UTC()
			createErr := q.CreateTracerouteResult(ctx, sqlc.CreateTracerouteResultParams{
				ProbeStorageID:     input.ProbeStorageID,
				CheckStorageID:     input.CheckStorageID,
				StartedAt:          startedAt,
				FinishedAt:         input.FinishedAt.UTC(),
				DurationMs:         input.DurationMs,
				Status:             sqlcTracerouteStatus(input.Status),
				ResolvedIp:         input.ResolvedIP,
				IpFamily:           sqlcIPFamily(input.IPFamily),
				DestinationReached: input.DestinationReached,
				HopCount:           input.HopCount,
				ErrorCode:          input.ErrorCode,
				ErrorMessage:       input.ErrorMessage,
			})
			if createErr != nil {
				return mapTracerouteResultWriteError(createErr)
			}
			for _, hop := range input.Hops {
				hopErr := q.CreateTracerouteResultHop(ctx, sqlc.CreateTracerouteResultHopParams{
					ProbeStorageID: input.ProbeStorageID,
					CheckStorageID: input.CheckStorageID,
					StartedAt:      startedAt,
					HopIndex:       hop.HopIndex,
					Address:        hop.Address,
					Hostname:       hop.Hostname,
					SentCount:      hop.SentCount,
					ReceivedCount:  hop.ReceivedCount,
					LossPercent:    hop.LossPercent,
					RttMinMs:       hop.RttMinMs,
					RttAvgMs:       hop.RttAvgMs,
					RttMedianMs:    hop.RttMedianMs,
					RttMaxMs:       hop.RttMaxMs,
					RttStddevMs:    hop.RttStddevMs,
					RttSamplesMs:   storageRTTSamples(hop.RttSamplesMs),
					ErrorCode:      hop.ErrorCode,
					ErrorMessage:   hop.ErrorMessage,
				})
				if hopErr != nil {
					return mapTracerouteResultWriteError(hopErr)
				}
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

func (r *TracerouteRepository) ListTracerouteRuns(ctx context.Context, input domaintraceroute.RunQuery) (domaintraceroute.RunResult, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgtracerouteTracer, "traceroute_results", "postgres.traceroute_results.runs", "SELECT", "SELECT traceroute runs")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domaintraceroute.RunResult{}, err
	}
	probeID, err := postgres.ParseUUID(input.ProbeID, domainprobe.ErrInvalidInput)
	if err != nil {
		return domaintraceroute.RunResult{}, err
	}
	checkID, err := postgres.ParseUUID(input.CheckID, domaincheck.ErrInvalidInput)
	if err != nil {
		return domaintraceroute.RunResult{}, err
	}

	storageIDs, err := r.queries.ResolveTracerouteRunStorageIDs(ctx, sqlc.ResolveTracerouteRunStorageIDsParams{
		CheckID:   checkID,
		ProjectID: projectID,
		ProbeID:   probeID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domaintraceroute.RunResult{}, domainprobe.ErrProbeNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return domaintraceroute.RunResult{}, err
	}

	rows, err := r.queries.ListTracerouteRunRows(ctx, sqlc.ListTracerouteRunRowsParams{
		ProbeStorageID:  storageIDs.ProbeStorageID,
		CheckStorageID:  storageIDs.CheckStorageID,
		StartedAtFrom:   input.From.UTC(),
		StartedAtTo:     input.To.UTC(),
		CursorStartedAt: optionalTime(input.Cursor),
		LimitCount:      input.Limit + 1,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domaintraceroute.RunResult{}, err
	}

	return mapRunRows(rows, input.Limit), nil
}

func (r *TracerouteRepository) ListTracerouteInsight(ctx context.Context, input domaintraceroute.InsightQuery) (domaintraceroute.InsightResult, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgtracerouteTracer, "traceroute_results", "postgres.traceroute_results.insight", "SELECT", "SELECT traceroute insight")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domaintraceroute.InsightResult{}, err
	}
	probeID, err := postgres.ParseUUID(input.ProbeID, domainprobe.ErrInvalidInput)
	if err != nil {
		return domaintraceroute.InsightResult{}, err
	}
	checkID, err := postgres.ParseUUID(input.CheckID, domaincheck.ErrInvalidInput)
	if err != nil {
		return domaintraceroute.InsightResult{}, err
	}

	storageIDs, err := r.queries.ResolveTracerouteRunStorageIDs(ctx, sqlc.ResolveTracerouteRunStorageIDsParams{
		CheckID:   checkID,
		ProjectID: projectID,
		ProbeID:   probeID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domaintraceroute.InsightResult{}, domainprobe.ErrProbeNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return domaintraceroute.InsightResult{}, err
	}

	startedAtFrom := input.From.UTC()
	startedAtTo := input.To.UTC()
	countParams := sqlc.CountTracerouteInsightPointsParams{
		ProbeStorageID: storageIDs.ProbeStorageID,
		CheckStorageID: storageIDs.CheckStorageID,
		StartedAtFrom:  startedAtFrom,
		StartedAtTo:    startedAtTo,
	}
	totalRuns, err := r.queries.CountTracerouteInsightPoints(ctx, countParams)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domaintraceroute.InsightResult{}, err
	}

	if totalRuns <= 200 {
		rows, rawErr := r.queries.ListTracerouteInsightRawRows(ctx, sqlc.ListTracerouteInsightRawRowsParams{
			ProbeStorageID: storageIDs.ProbeStorageID,
			CheckStorageID: storageIDs.CheckStorageID,
			StartedAtFrom:  startedAtFrom,
			StartedAtTo:    startedAtTo,
		})
		if rawErr != nil {
			postgres.RecordDBSpanError(span, rawErr)
			return domaintraceroute.InsightResult{}, rawErr
		}
		return domaintraceroute.InsightResult{
			Points:     mapRawInsightRows(rows),
			Resolution: domaintraceroute.InsightResolutionRaw,
			TotalRuns:  totalRuns,
		}, nil
	}

	rows, bucketErr := r.queries.ListTracerouteInsightBucketRows(ctx, sqlc.ListTracerouteInsightBucketRowsParams{
		ProbeStorageID: storageIDs.ProbeStorageID,
		CheckStorageID: storageIDs.CheckStorageID,
		StartedAtFrom:  startedAtFrom,
		StartedAtTo:    startedAtTo,
		MaxDataPoints:  float64(input.MaxDataPoints),
	})
	if bucketErr != nil {
		postgres.RecordDBSpanError(span, bucketErr)
		return domaintraceroute.InsightResult{}, bucketErr
	}
	return domaintraceroute.InsightResult{
		Points:     mapBucketInsightRows(rows),
		Resolution: domaintraceroute.InsightResolutionBucket,
		TotalRuns:  totalRuns,
	}, nil
}

func (r *TracerouteRepository) ListTracerouteTopologyRuns(ctx context.Context, input domaintraceroute.TopologyQuery) (domaintraceroute.TopologyRunResult, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgtracerouteTracer, "traceroute_results", "postgres.traceroute_results.topology", "SELECT", "SELECT traceroute topology")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domaintraceroute.TopologyRunResult{}, err
	}
	probeID, err := optionalUUID(input.ProbeID, domainprobe.ErrInvalidInput)
	if err != nil {
		return domaintraceroute.TopologyRunResult{}, err
	}
	checkID, err := optionalUUID(input.CheckID, domaincheck.ErrInvalidInput)
	if err != nil {
		return domaintraceroute.TopologyRunResult{}, err
	}

	rows, err := r.queries.ListTracerouteTopologyRows(ctx, sqlc.ListTracerouteTopologyRowsParams{
		ProjectID:     projectID,
		ProbeID:       probeID,
		CheckID:       checkID,
		StartedAtFrom: input.From.UTC(),
		StartedAtTo:   input.To.UTC(),
		LimitCount:    input.Limit,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domaintraceroute.TopologyRunResult{}, err
	}

	return mapTopologyRows(rows), nil
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
