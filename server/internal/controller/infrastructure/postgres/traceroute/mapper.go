package pgtraceroute

import (
	"time"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

func sqlcTracerouteStatus(value domaintraceroute.Status) sqlc.TracerouteStatus {
	return sqlc.TracerouteStatus(value)
}

func sqlcIPFamily(value *domainnetwork.IPFamily) *sqlc.IpFamily {
	if value == nil {
		return nil
	}

	ipFamily := sqlc.IpFamily(*value)
	return &ipFamily
}

func mapIPFamily(value *sqlc.IpFamily) *domainnetwork.IPFamily {
	if value == nil {
		return nil
	}

	ipFamily := domainnetwork.IPFamily(*value)
	return &ipFamily
}

func optionalTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := value.UTC()
	return &copied
}

func storageRTTSamples(samples []float64) []float64 {
	return append([]float64{}, samples...)
}

func mapRunRows(rows []sqlc.ListTracerouteRunRowsRow, limit int32) domaintraceroute.RunResult {
	limitCount := int(limit)
	runIndex := make(map[time.Time]int)
	runs := make([]domaintraceroute.Run, 0)
	for _, row := range rows {
		startedAt := row.StartedAt.UTC()
		index, ok := runIndex[startedAt]
		if !ok {
			if len(runs) >= limitCount+1 {
				continue
			}
			index = len(runs)
			runIndex[startedAt] = index
			runs = append(runs, domaintraceroute.Run{
				StartedAt:          startedAt,
				FinishedAt:         row.FinishedAt.UTC(),
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
		if row.HopIndex > 0 {
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
		startedAt := row.StartedAt.UTC()
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
		if row.HopIndex > 0 {
			runs[index].Hops = append(runs[index].Hops, domaintraceroute.TopologyHop{
				HopIndex:    row.HopIndex,
				Address:     row.Address,
				Hostname:    row.Hostname,
				LossPercent: row.LossPercent,
				RttAvgMs:    row.RttAvgMs,
			})
		}
	}

	return domaintraceroute.TopologyRunResult{Runs: runs}
}

func mapRawInsightRows(rows []sqlc.ListTracerouteInsightRawRowsRow) []domaintraceroute.InsightPoint {
	values := make([]domaintraceroute.InsightPoint, 0, len(rows))
	for _, row := range rows {
		runStartedAt := row.RunStartedAt.UTC()
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
		HopIndex:      row.HopIndex,
		Address:       row.Address,
		Hostname:      row.Hostname,
		SentCount:     row.SentCount,
		ReceivedCount: row.ReceivedCount,
		LossPercent:   row.LossPercent,
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

func floatPtrIf(count int64, value float64) *float64 {
	if count == 0 {
		return nil
	}
	copied := value
	return &copied
}
