package proberuntime

import (
	"net/netip"
	"time"

	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

type normalizedPingTiming struct {
	startedAt  time.Time
	finishedAt time.Time
	durationMs int32
}

type normalizedPingCounts struct {
	status        domainping.Status
	sentCount     int32
	receivedCount int32
	lossPercent   float64
}

type normalizedPingRTT struct {
	min     *float64
	avg     *float64
	median  *float64
	max     *float64
	stddev  *float64
	samples []float64
}

type normalizedPingMetadata struct {
	ipFamily     *domainnetwork.IPFamily
	errorCode    *string
	errorMessage *string
}

func normalizePingResult(input PingResultInput, fieldPrefix string) (domainping.ResultStorageInput, error) {
	timing, err := normalizePingTiming(input, fieldPrefix)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	counts, err := normalizePingCounts(input, fieldPrefix)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	rtt, err := normalizePingRTT(input, fieldPrefix)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	metadata, err := normalizePingMetadata(input, fieldPrefix)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}

	return domainping.ResultStorageInput{
		StartedAt:     timing.startedAt,
		FinishedAt:    timing.finishedAt,
		DurationMs:    timing.durationMs,
		Status:        counts.status,
		SentCount:     counts.sentCount,
		ReceivedCount: counts.receivedCount,
		LossPercent:   counts.lossPercent,
		RttMinMs:      rtt.min,
		RttAvgMs:      rtt.avg,
		RttMedianMs:   rtt.median,
		RttMaxMs:      rtt.max,
		RttStddevMs:   rtt.stddev,
		RttSamplesMs:  rtt.samples,
		ResolvedIP:    cloneAddr(input.ResolvedIP),
		IPFamily:      metadata.ipFamily,
		ErrorCode:     metadata.errorCode,
		ErrorMessage:  metadata.errorMessage,
	}, nil
}

func normalizePingTiming(input PingResultInput, fieldPrefix string) (normalizedPingTiming, error) {
	startedAt, err := domainping.VNResultTime(input.StartedAt)
	if err != nil {
		return normalizedPingTiming{}, invalidRuntimeField(resultField(fieldPrefix, "startedAt"), err.Error(), input.StartedAt)
	}
	finishedAt, err := domainping.VNResultTime(input.FinishedAt)
	if err != nil {
		return normalizedPingTiming{}, invalidRuntimeField(resultField(fieldPrefix, "finishedAt"), err.Error(), input.FinishedAt)
	}
	if finishedAt.Before(startedAt) {
		return normalizedPingTiming{}, invalidRuntimeField(resultField(fieldPrefix, "finishedAt"), "must be greater than or equal to startedAt", input.FinishedAt)
	}
	durationMs, err := domainping.VNResultDurationMs(input.DurationMs)
	if err != nil {
		return normalizedPingTiming{}, invalidRuntimeField(resultField(fieldPrefix, "durationMs"), err.Error(), input.DurationMs)
	}

	return normalizedPingTiming{
		startedAt:  startedAt,
		finishedAt: finishedAt,
		durationMs: durationMs,
	}, nil
}

func normalizePingCounts(input PingResultInput, fieldPrefix string) (normalizedPingCounts, error) {
	status, err := domainping.VNResultStatus(domainping.Status(input.Status))
	if err != nil {
		return normalizedPingCounts{}, invalidRuntimeField(resultField(fieldPrefix, "status"), err.Error(), input.Status)
	}
	sentCount, err := domainping.VNResultSentCount(input.SentCount)
	if err != nil {
		return normalizedPingCounts{}, invalidRuntimeField(resultField(fieldPrefix, "sentCount"), err.Error(), input.SentCount)
	}
	receivedCount, err := domainping.VNResultReceivedCount(input.ReceivedCount, sentCount)
	if err != nil {
		return normalizedPingCounts{}, invalidRuntimeField(resultField(fieldPrefix, "receivedCount"), err.Error(), input.ReceivedCount)
	}
	lossPercent, err := domainping.VNResultLossPercent(input.LossPercent)
	if err != nil {
		return normalizedPingCounts{}, invalidRuntimeField(resultField(fieldPrefix, "lossPercent"), err.Error(), input.LossPercent)
	}

	return normalizedPingCounts{
		status:        status,
		sentCount:     sentCount,
		receivedCount: receivedCount,
		lossPercent:   lossPercent,
	}, nil
}

func normalizePingRTT(input PingResultInput, fieldPrefix string) (normalizedPingRTT, error) {
	rttMin, err := normalizeOptionalResultRTT(input.RttMinMs, resultField(fieldPrefix, "rttMinMs"))
	if err != nil {
		return normalizedPingRTT{}, err
	}
	rttAvg, err := normalizeOptionalResultRTT(input.RttAvgMs, resultField(fieldPrefix, "rttAvgMs"))
	if err != nil {
		return normalizedPingRTT{}, err
	}
	rttMedian, err := normalizeOptionalResultRTT(input.RttMedianMs, resultField(fieldPrefix, "rttMedianMs"))
	if err != nil {
		return normalizedPingRTT{}, err
	}
	rttMax, err := normalizeOptionalResultRTT(input.RttMaxMs, resultField(fieldPrefix, "rttMaxMs"))
	if err != nil {
		return normalizedPingRTT{}, err
	}
	rttStddev, err := normalizeOptionalResultRTT(input.RttStddevMs, resultField(fieldPrefix, "rttStddevMs"))
	if err != nil {
		return normalizedPingRTT{}, err
	}
	err = validateRTTOrder(rttMin, rttAvg, rttMax, fieldPrefix)
	if err != nil {
		return normalizedPingRTT{}, err
	}
	rttSamples, err := domainping.VNResultRTTSamples(input.RttSamplesMs)
	if err != nil {
		return normalizedPingRTT{}, invalidRuntimeField(resultField(fieldPrefix, "rttSamplesMs"), err.Error(), input.RttSamplesMs)
	}

	return normalizedPingRTT{
		min:     rttMin,
		avg:     rttAvg,
		median:  rttMedian,
		max:     rttMax,
		stddev:  rttStddev,
		samples: rttSamples,
	}, nil
}

func normalizePingMetadata(input PingResultInput, fieldPrefix string) (normalizedPingMetadata, error) {
	ipFamily, err := domainnetwork.ParseOptionalIPFamily(input.IPFamily)
	if err != nil {
		return normalizedPingMetadata{}, invalidRuntimeField(resultField(fieldPrefix, "ipFamily"), `must be "inet" or "inet6"`, input.IPFamily)
	}
	errorCode, err := normalizeOptionalResultText(input.ErrorCode, resultField(fieldPrefix, "errorCode"))
	if err != nil {
		return normalizedPingMetadata{}, err
	}
	errorMessage, err := normalizeOptionalResultText(input.ErrorMessage, resultField(fieldPrefix, "errorMessage"))
	if err != nil {
		return normalizedPingMetadata{}, err
	}

	return normalizedPingMetadata{
		ipFamily:     ipFamily,
		errorCode:    errorCode,
		errorMessage: errorMessage,
	}, nil
}

func normalizeOptionalResultRTT(input *float64, field string) (*float64, error) {
	value, err := domainping.VNResultOptionalRTT(input)
	if err != nil {
		return nil, invalidRuntimeField(field, err.Error(), input)
	}
	return value, nil
}

func normalizeOptionalResultText(input *string, field string) (*string, error) {
	value, err := domainping.VNResultOptionalText(input)
	if err != nil {
		return nil, invalidRuntimeField(field, err.Error(), input)
	}
	return value, nil
}

func validateRTTOrder(minValue, avgValue, maxValue *float64, fieldPrefix string) error {
	if minValue != nil && maxValue != nil && *minValue > *maxValue {
		return invalidRuntimeField(resultField(fieldPrefix, "rttMinMs"), "must be less than or equal to rttMaxMs", minValue)
	}
	if minValue != nil && avgValue != nil && *minValue > *avgValue {
		return invalidRuntimeField(resultField(fieldPrefix, "rttMinMs"), "must be less than or equal to rttAvgMs", minValue)
	}
	if avgValue != nil && maxValue != nil && *avgValue > *maxValue {
		return invalidRuntimeField(resultField(fieldPrefix, "rttAvgMs"), "must be less than or equal to rttMaxMs", avgValue)
	}

	return nil
}

func resultField(prefix, field string) string {
	return prefix + "." + field
}

func cloneAddr(value *netip.Addr) *netip.Addr {
	if value == nil {
		return nil
	}

	copied := *value
	return &copied
}
