package pgtraceroute

import (
	"context"
	"errors"

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
			startedAt := timestamptz(input.StartedAt)
			createErr := q.CreateTracerouteResult(ctx, sqlc.CreateTracerouteResultParams{
				ProbeStorageID:     input.ProbeStorageID,
				CheckStorageID:     input.CheckStorageID,
				StartedAt:          startedAt,
				FinishedAt:         timestamptz(input.FinishedAt),
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
		StartedAtFrom:   timestamptz(input.From),
		StartedAtTo:     timestamptz(input.To),
		CursorStartedAt: optionalTimestamptz(input.Cursor),
		LimitCount:      input.Limit + 1,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domaintraceroute.RunResult{}, err
	}

	return mapRunRows(rows, input.Limit), nil
}
