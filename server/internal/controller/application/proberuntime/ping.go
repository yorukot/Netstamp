package proberuntime

import (
	"net/netip"
	"time"

	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

type normalizedResultTiming struct {
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

type normalizedResultRTT struct {
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
	timing, err := normalizeResultTiming(
		input.StartedAt,
		input.FinishedAt,
		input.DurationMs,
		fieldPrefix,
		domainping.VNResultDurationMs,
	)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	counts, err := normalizePingCounts(input, fieldPrefix)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	rtt, err := normalizeResultRTT(
		input.RttMinMs,
		input.RttAvgMs,
		input.RttMedianMs,
		input.RttMaxMs,
		input.RttStddevMs,
		input.RttSamplesMs,
		fieldPrefix,
		normalizeOptionalResultRTT,
		domainping.VNResultRTTSamples,
	)
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

func normalizeResultTiming(startedAtInput, finishedAtInput time.Time, durationMsInput int32, fieldPrefix string, validateDuration func(int32) (int32, error)) (normalizedResultTiming, error) {
	startedAt, err := normalizeResultTimestamp(startedAtInput, resultField(fieldPrefix, "startedAt"))
	if err != nil {
		return normalizedResultTiming{}, err
	}
	finishedAt, err := normalizeResultTimestamp(finishedAtInput, resultField(fieldPrefix, "finishedAt"))
	if err != nil {
		return normalizedResultTiming{}, err
	}
	if finishedAt.Before(startedAt) {
		return normalizedResultTiming{}, invalidRuntimeField(resultField(fieldPrefix, "finishedAt"), "must be greater than or equal to startedAt", finishedAtInput)
	}
	durationMs, err := validateDuration(durationMsInput)
	if err != nil {
		return normalizedResultTiming{}, invalidRuntimeField(resultField(fieldPrefix, "durationMs"), err.Error(), durationMsInput)
	}

	return normalizedResultTiming{
		startedAt:  startedAt,
		finishedAt: finishedAt,
		durationMs: durationMs,
	}, nil
}

func normalizeResultTimestamp(input time.Time, field string) (time.Time, error) {
	if input.IsZero() {
		return time.Time{}, invalidRuntimeField(field, "must be provided", input)
	}
	return input.UTC(), nil
}

func normalizeResultRTT(rttMinInput, rttAvgInput, rttMedianInput, rttMaxInput, rttStddevInput *float64, samplesInput []float64, fieldPrefix string, normalizeOptional func(*float64, string) (*float64, error), validateSamples func([]float64) ([]float64, error)) (normalizedResultRTT, error) {
	rttMin, err := normalizeOptional(rttMinInput, resultField(fieldPrefix, "rttMinMs"))
	if err != nil {
		return normalizedResultRTT{}, err
	}
	rttAvg, err := normalizeOptional(rttAvgInput, resultField(fieldPrefix, "rttAvgMs"))
	if err != nil {
		return normalizedResultRTT{}, err
	}
	rttMedian, err := normalizeOptional(rttMedianInput, resultField(fieldPrefix, "rttMedianMs"))
	if err != nil {
		return normalizedResultRTT{}, err
	}
	rttMax, err := normalizeOptional(rttMaxInput, resultField(fieldPrefix, "rttMaxMs"))
	if err != nil {
		return normalizedResultRTT{}, err
	}
	rttStddev, err := normalizeOptional(rttStddevInput, resultField(fieldPrefix, "rttStddevMs"))
	if err != nil {
		return normalizedResultRTT{}, err
	}
	err = validateRTTOrder(rttMin, rttAvg, rttMax, fieldPrefix)
	if err != nil {
		return normalizedResultRTT{}, err
	}
	rttSamples, err := validateSamples(samplesInput)
	if err != nil {
		return normalizedResultRTT{}, invalidRuntimeField(resultField(fieldPrefix, "rttSamplesMs"), err.Error(), samplesInput)
	}

	return normalizedResultRTT{
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
