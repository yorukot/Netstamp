package proberuntime

import (
	"fmt"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type normalizedTracerouteMetadata struct {
	ipFamily     *domainnetwork.IPFamily
	errorCode    *string
	errorMessage *string
}

type normalizedTracerouteHopCounts struct {
	sentCount     int32
	receivedCount int32
	lossPercent   float64
}

func normalizeTracerouteResult(input TracerouteResultInput, fieldPrefix string) (domaintraceroute.ResultStorageInput, error) {
	var validation appvalidation.Collector

	timing, err := normalizeResultTiming(
		input.StartedAt,
		input.FinishedAt,
		input.DurationMs,
		fieldPrefix,
		domaintraceroute.VNResultDurationMs,
	)
	if err != nil {
		if !validation.AddValidation(err) {
			return domaintraceroute.ResultStorageInput{}, err
		}
	}
	status, err := normalizeTracerouteStatus(input.Status, fieldPrefix)
	if err != nil {
		if !validation.AddValidation(err) {
			return domaintraceroute.ResultStorageInput{}, err
		}
	}
	hopCount, err := domaintraceroute.VNResultHopCount(input.HopCount)
	if err != nil {
		validation.AddError(resultField(fieldPrefix, "hopCount"), err, input.HopCount)
	}
	metadata, err := normalizeTracerouteMetadata(input, fieldPrefix)
	if err != nil {
		if !validation.AddValidation(err) {
			return domaintraceroute.ResultStorageInput{}, err
		}
	}
	hops, err := normalizeTracerouteHops(input.Hops, fieldPrefix)
	if err != nil {
		if !validation.AddValidation(err) {
			return domaintraceroute.ResultStorageInput{}, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return domaintraceroute.ResultStorageInput{}, err
	}

	return domaintraceroute.ResultStorageInput{
		StartedAt:          timing.startedAt,
		FinishedAt:         timing.finishedAt,
		DurationMs:         timing.durationMs,
		Status:             status,
		ResolvedIP:         cloneAddr(input.ResolvedIP),
		IPFamily:           metadata.ipFamily,
		DestinationReached: input.DestinationReached,
		HopCount:           hopCount,
		ErrorCode:          metadata.errorCode,
		ErrorMessage:       metadata.errorMessage,
		Hops:               hops,
	}, nil
}

func normalizeTracerouteStatus(input, fieldPrefix string) (domaintraceroute.Status, error) {
	status, err := domaintraceroute.VNResultStatus(domaintraceroute.Status(input))
	if err != nil {
		return "", invalidRuntimeField(resultField(fieldPrefix, "status"), err.Error(), input)
	}
	return status, nil
}

func normalizeTracerouteMetadata(input TracerouteResultInput, fieldPrefix string) (normalizedTracerouteMetadata, error) {
	metadata, err := normalizeResultMetadata(input.IPFamily, input.ErrorCode, input.ErrorMessage, fieldPrefix, normalizeOptionalTracerouteText)
	if err != nil {
		return normalizedTracerouteMetadata{}, err
	}

	return normalizedTracerouteMetadata(metadata), nil
}

func normalizeTracerouteHops(inputs []TracerouteHopInput, fieldPrefix string) ([]domaintraceroute.HopStorageInput, error) {
	var validation appvalidation.Collector

	hops := make([]domaintraceroute.HopStorageInput, 0, len(inputs))
	seen := map[int32]struct{}{}
	for i, input := range inputs {
		prefix := resultField(fieldPrefix, fmt.Sprintf("hops[%d]", i))
		hop, err := normalizeTracerouteHop(input, prefix)
		if err != nil {
			if !validation.AddValidation(err) {
				return nil, err
			}
			continue
		}
		if _, ok := seen[hop.HopIndex]; ok {
			validation.Add(resultField(prefix, "hopIndex"), "duplicate hop index", hop.HopIndex)
			continue
		}
		seen[hop.HopIndex] = struct{}{}
		hops = append(hops, hop)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return nil, err
	}

	return hops, nil
}

func normalizeTracerouteHop(input TracerouteHopInput, fieldPrefix string) (domaintraceroute.HopStorageInput, error) {
	var validation appvalidation.Collector

	hopIndex, err := domaintraceroute.VNResultHopIndex(input.HopIndex)
	if err != nil {
		validation.AddError(resultField(fieldPrefix, "hopIndex"), err, input.HopIndex)
	}
	counts, err := normalizeTracerouteHopCounts(input, fieldPrefix)
	if err != nil {
		if !validation.AddValidation(err) {
			return domaintraceroute.HopStorageInput{}, err
		}
	}
	rtt, err := normalizeTracerouteHopRTT(input, fieldPrefix)
	if err != nil {
		if !validation.AddValidation(err) {
			return domaintraceroute.HopStorageInput{}, err
		}
	}
	hostname, err := normalizeOptionalTracerouteText(input.Hostname, resultField(fieldPrefix, "hostname"))
	if err != nil {
		if !validation.AddValidation(err) {
			return domaintraceroute.HopStorageInput{}, err
		}
	}
	errorCode, err := normalizeOptionalTracerouteText(input.ErrorCode, resultField(fieldPrefix, "errorCode"))
	if err != nil {
		if !validation.AddValidation(err) {
			return domaintraceroute.HopStorageInput{}, err
		}
	}
	errorMessage, err := normalizeOptionalTracerouteText(input.ErrorMessage, resultField(fieldPrefix, "errorMessage"))
	if err != nil {
		if !validation.AddValidation(err) {
			return domaintraceroute.HopStorageInput{}, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return domaintraceroute.HopStorageInput{}, err
	}

	return domaintraceroute.HopStorageInput{
		HopIndex:      hopIndex,
		Address:       cloneAddr(input.Address),
		Hostname:      hostname,
		SentCount:     counts.sentCount,
		ReceivedCount: counts.receivedCount,
		LossPercent:   counts.lossPercent,
		RttMinMs:      rtt.min,
		RttAvgMs:      rtt.avg,
		RttMedianMs:   rtt.median,
		RttMaxMs:      rtt.max,
		RttStddevMs:   rtt.stddev,
		RttSamplesMs:  rtt.samples,
		ErrorCode:     errorCode,
		ErrorMessage:  errorMessage,
	}, nil
}

func normalizeTracerouteHopCounts(input TracerouteHopInput, fieldPrefix string) (normalizedTracerouteHopCounts, error) {
	var validation appvalidation.Collector

	sentCount, err := domaintraceroute.VNResultSentCount(input.SentCount)
	if err != nil {
		validation.AddError(resultField(fieldPrefix, "sentCount"), err, input.SentCount)
	}
	var receivedCount int32
	if err == nil {
		receivedCount, err = domaintraceroute.VNResultReceivedCount(input.ReceivedCount, sentCount)
		if err != nil {
			validation.AddError(resultField(fieldPrefix, "receivedCount"), err, input.ReceivedCount)
		}
	}
	lossPercent, err := domaintraceroute.VNResultLossPercent(input.LossPercent)
	if err != nil {
		validation.AddError(resultField(fieldPrefix, "lossPercent"), err, input.LossPercent)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedTracerouteHopCounts{}, err
	}

	return normalizedTracerouteHopCounts{
		sentCount:     sentCount,
		receivedCount: receivedCount,
		lossPercent:   lossPercent,
	}, nil
}

func normalizeTracerouteHopRTT(input TracerouteHopInput, fieldPrefix string) (normalizedResultRTT, error) {
	return normalizeResultRTT(
		input.RttMinMs,
		input.RttAvgMs,
		input.RttMedianMs,
		input.RttMaxMs,
		input.RttStddevMs,
		input.RttSamplesMs,
		fieldPrefix,
		normalizeTracerouteOptionalRTT,
		domaintraceroute.VNResultRTTSamples,
	)
}

func normalizeTracerouteOptionalRTT(input *float64, field string) (*float64, error) {
	value, err := domaintraceroute.VNResultOptionalRTT(input)
	if err != nil {
		return nil, invalidRuntimeField(field, err.Error(), input)
	}
	return value, nil
}

func normalizeOptionalTracerouteText(input *string, field string) (*string, error) {
	value, err := domaintraceroute.VNResultOptionalText(input)
	if err != nil {
		return nil, invalidRuntimeField(field, err.Error(), input)
	}
	return value, nil
}
