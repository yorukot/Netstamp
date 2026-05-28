package pgtraceroute

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

func sqlcTracerouteStatus(value domaintraceroute.Status) sqlc.TracerouteStatus {
	return sqlc.TracerouteStatus(value)
}

func sqlcIPFamily(value *domainnetwork.IPFamily) sqlc.NullIpFamily {
	if value == nil {
		return sqlc.NullIpFamily{}
	}

	return sqlc.NullIpFamily{
		IpFamily: sqlc.IpFamily(*value),
		Valid:    true,
	}
}

func mapIPFamily(value sqlc.NullIpFamily) *domainnetwork.IPFamily {
	if !value.Valid {
		return nil
	}

	ipFamily := domainnetwork.IPFamily(value.IpFamily)
	return &ipFamily
}

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func optionalTimestamptz(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}

	return timestamptz(*value)
}

func storageRTTSamples(samples []float64) []float64 {
	return append([]float64{}, samples...)
}

func mapRunRows(rows []sqlc.ListTracerouteRunRowsRow, limit int32) domaintraceroute.RunResult {
	limitCount := int(limit)
	runIndex := make(map[time.Time]int)
	runs := make([]domaintraceroute.Run, 0)
	for _, row := range rows {
		startedAt := row.StartedAt.Time.UTC()
		index, ok := runIndex[startedAt]
		if !ok {
			if len(runs) >= limitCount+1 {
				continue
			}
			index = len(runs)
			runIndex[startedAt] = index
			runs = append(runs, domaintraceroute.Run{
				StartedAt:          startedAt,
				FinishedAt:         row.FinishedAt.Time.UTC(),
				DurationMs:         row.DurationMs,
				Status:             domaintraceroute.Status(row.Status),
				ResolvedIP:         row.ResolvedIp,
				IPFamily:           mapIPFamily(row.IpFamily),
				DestinationReached: row.DestinationReached,
				HopCount:           row.HopCount,
				ErrorCode:          row.ErrorCode,
				ErrorMessage:       row.ErrorMessage,
				Hops:               []domaintraceroute.Hop{},
			})
		}
		if row.HopIndex != nil {
			runs[index].Hops = append(runs[index].Hops, mapHop(row))
		}
	}

	var nextCursor *time.Time
	if len(runs) > limitCount {
		cursor := runs[limitCount-1].StartedAt
		nextCursor = &cursor
		runs = runs[:limitCount]
	}

	return domaintraceroute.RunResult{
		Runs:       runs,
		NextCursor: nextCursor,
	}
}

func mapTopologyRows(rows []sqlc.ListTracerouteTopologyRowsRow) domaintraceroute.TopologyRunResult {
	type runKey struct {
		startedAt time.Time
		probeID   string
		checkID   string
	}

	runIndex := make(map[runKey]int)
	runs := make([]domaintraceroute.TopologyRun, 0)
	for _, row := range rows {
		startedAt := row.StartedAt.Time.UTC()
		probeID := row.ProbePublicID.String()
		checkID := row.CheckPublicID.String()
		key := runKey{startedAt: startedAt, probeID: probeID, checkID: checkID}
		index, ok := runIndex[key]
		if !ok {
			index = len(runs)
			runIndex[key] = index
			runs = append(runs, domaintraceroute.TopologyRun{
				StartedAt:  startedAt,
				ProbeID:    probeID,
				ProbeName:  row.ProbeName,
				CheckID:    checkID,
				CheckName:  row.CheckName,
				Target:     row.CheckTarget,
				ResolvedIP: row.ResolvedIp,
				Hops:       []domaintraceroute.TopologyHop{},
			})
		}
		if row.HopIndex != nil {
			runs[index].Hops = append(runs[index].Hops, domaintraceroute.TopologyHop{
				HopIndex:    derefInt32(row.HopIndex),
				Address:     row.Address,
				Hostname:    row.Hostname,
				LossPercent: derefFloat64(row.LossPercent),
				RttAvgMs:    row.RttAvgMs,
			})
		}
	}

	return domaintraceroute.TopologyRunResult{Runs: runs}
}

func mapRawInsightRows(rows []sqlc.ListTracerouteInsightRawRowsRow) []domaintraceroute.InsightPoint {
	values := make([]domaintraceroute.InsightPoint, 0, len(rows))
	for _, row := range rows {
		runStartedAt := row.RunStartedAt.Time.UTC()
		values = append(values, domaintraceroute.InsightPoint{
			Timestamp:          time.UnixMilli(row.BucketMs).UTC(),
			BucketFrom:         time.UnixMilli(row.BucketFromMs).UTC(),
			BucketTo:           time.UnixMilli(row.BucketToMs).UTC(),
			RunStartedAt:       &runStartedAt,
			ResultCount:        row.ResultCount,
			FinalRttAvgMs:      floatPtrIf(row.FinalRttValueCount, row.FinalRttAvgMs),
			FinalLossPercent:   floatPtrIf(row.FinalLossValueCount, row.FinalLossPercent),
			HasLoss:            row.HasLoss,
			HasRouteChange:     row.HasRouteChange,
			DestinationReached: row.DestinationReached,
		})
	}
	return values
}

func mapBucketInsightRows(rows []sqlc.ListTracerouteInsightBucketRowsRow) []domaintraceroute.InsightPoint {
	values := make([]domaintraceroute.InsightPoint, 0, len(rows))
	for _, row := range rows {
		values = append(values, domaintraceroute.InsightPoint{
			Timestamp:          time.UnixMilli(row.BucketMs).UTC(),
			BucketFrom:         time.UnixMilli(row.BucketFromMs).UTC(),
			BucketTo:           time.UnixMilli(row.BucketToMs).UTC(),
			ResultCount:        row.ResultCount,
			FinalRttAvgMs:      floatPtrIf(row.FinalRttValueCount, row.FinalRttAvgMs),
			FinalLossPercent:   floatPtrIf(row.FinalLossValueCount, row.FinalLossPercent),
			HasLoss:            row.HasLoss,
			HasRouteChange:     row.HasRouteChange,
			DestinationReached: row.DestinationReached,
		})
	}
	return values
}

func mapHop(row sqlc.ListTracerouteRunRowsRow) domaintraceroute.Hop {
	return domaintraceroute.Hop{
		HopIndex:      derefInt32(row.HopIndex),
		Address:       row.Address,
		Hostname:      row.Hostname,
		SentCount:     derefInt32(row.SentCount),
		ReceivedCount: derefInt32(row.ReceivedCount),
		LossPercent:   derefFloat64(row.LossPercent),
		RttMinMs:      row.RttMinMs,
		RttAvgMs:      row.RttAvgMs,
		RttMedianMs:   row.RttMedianMs,
		RttMaxMs:      row.RttMaxMs,
		RttStddevMs:   row.RttStddevMs,
		RttSamplesMs:  append([]float64(nil), row.RttSamplesMs...),
		ErrorCode:     row.HopErrorCode,
		ErrorMessage:  row.HopErrorMessage,
	}
}

func derefInt32(value *int32) int32 {
	if value == nil {
		return 0
	}
	return *value
}

func derefFloat64(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func floatPtrIf(count int64, value float64) *float64 {
	if count == 0 {
		return nil
	}
	copied := value
	return &copied
}
