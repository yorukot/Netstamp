package proberuntime

import (
	"net/netip"
	"time"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
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

type normalizedResultMetadata struct {
	ipFamily     *domainnetwork.IPFamily
	errorCode    *string
	errorMessage *string
}

func normalizePingResult(input PingResultInput, fieldPrefix string) (domainping.ResultStorageInput, error) {
	var validation appvalidation.Collector

	timing, err := normalizeResultTiming(
		input.StartedAt,
		input.FinishedAt,
		input.DurationMs,
		fieldPrefix,
		domainping.VNResultDurationMs,
	)
	if err != nil {
		if !validation.AddValidation(err) {
			return domainping.ResultStorageInput{}, err
		}
	}
	counts, err := normalizePingCounts(input, fieldPrefix)
	if err != nil {
		if !validation.AddValidation(err) {
			return domainping.ResultStorageInput{}, err
		}
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
		if !validation.AddValidation(err) {
			return domainping.ResultStorageInput{}, err
		}
	}
	metadata, err := normalizePingMetadata(input, fieldPrefix)
	if err != nil {
		if !validation.AddValidation(err) {
			return domainping.ResultStorageInput{}, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
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
	var validation appvalidation.Collector

	status, err := domainping.VNResultStatus(domainping.Status(input.Status))
	if err != nil {
		validation.AddError(resultField(fieldPrefix, "status"), err, input.Status)
	}
	sentCount, err := domainping.VNResultSentCount(input.SentCount)
	if err != nil {
		validation.AddError(resultField(fieldPrefix, "sentCount"), err, input.SentCount)
	}
	var receivedCount int32
	if err == nil {
		receivedCount, err = domainping.VNResultReceivedCount(input.ReceivedCount, sentCount)
		if err != nil {
			validation.AddError(resultField(fieldPrefix, "receivedCount"), err, input.ReceivedCount)
		}
	}
	lossPercent, err := domainping.VNResultLossPercent(input.LossPercent)
	if err != nil {
		validation.AddError(resultField(fieldPrefix, "lossPercent"), err, input.LossPercent)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedPingCounts{}, err
	}

	return normalizedPingCounts{
		status:        status,
		sentCount:     sentCount,
		receivedCount: receivedCount,
		lossPercent:   lossPercent,
	}, nil
}

func normalizeResultTiming(startedAtInput, finishedAtInput time.Time, durationMsInput int32, fieldPrefix string, validateDuration func(int32) (int32, error)) (normalizedResultTiming, error) {
	var validation appvalidation.Collector

	startedAt, err := normalizeResultTimestamp(startedAtInput, resultField(fieldPrefix, "startedAt"))
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedResultTiming{}, err
		}
	}
	startedAtValid := err == nil
	finishedAt, err := normalizeResultTimestamp(finishedAtInput, resultField(fieldPrefix, "finishedAt"))
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedResultTiming{}, err
		}
	}
	finishedAtValid := err == nil
	if startedAtValid && finishedAtValid && finishedAt.Before(startedAt) {
		validation.Add(resultField(fieldPrefix, "finishedAt"), "must be greater than or equal to startedAt", finishedAtInput)
	}
	durationMs, err := validateDuration(durationMsInput)
	if err != nil {
		validation.AddError(resultField(fieldPrefix, "durationMs"), err, durationMsInput)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedResultTiming{}, err
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
	var validation appvalidation.Collector

	rttMin, err := normalizeOptional(rttMinInput, resultField(fieldPrefix, "rttMinMs"))
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedResultRTT{}, err
		}
	}
	rttAvg, err := normalizeOptional(rttAvgInput, resultField(fieldPrefix, "rttAvgMs"))
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedResultRTT{}, err
		}
	}
	rttMedian, err := normalizeOptional(rttMedianInput, resultField(fieldPrefix, "rttMedianMs"))
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedResultRTT{}, err
		}
	}
	rttMax, err := normalizeOptional(rttMaxInput, resultField(fieldPrefix, "rttMaxMs"))
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedResultRTT{}, err
		}
	}
	rttStddev, err := normalizeOptional(rttStddevInput, resultField(fieldPrefix, "rttStddevMs"))
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedResultRTT{}, err
		}
	}
	err = validateRTTOrder(rttMin, rttAvg, rttMax, fieldPrefix)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedResultRTT{}, err
		}
	}
	rttSamples, err := validateSamples(samplesInput)
	if err != nil {
		validation.AddError(resultField(fieldPrefix, "rttSamplesMs"), err, samplesInput)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedResultRTT{}, err
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
	metadata, err := normalizeResultMetadata(input.IPFamily, input.ErrorCode, input.ErrorMessage, fieldPrefix, normalizeOptionalResultText)
	if err != nil {
		return normalizedPingMetadata{}, err
	}

	return normalizedPingMetadata(metadata), nil
}

func normalizeResultMetadata(ipFamilyInput, errorCodeInput, errorMessageInput *string, fieldPrefix string, normalizeText func(*string, string) (*string, error)) (normalizedResultMetadata, error) {
	var validation appvalidation.Collector

	ipFamily, err := domainnetwork.ParseOptionalIPFamily(ipFamilyInput)
	if err != nil {
		validation.Add(resultField(fieldPrefix, "ipFamily"), `must be "inet" or "inet6"`, ipFamilyInput)
	}
	errorCode, err := normalizeText(errorCodeInput, resultField(fieldPrefix, "errorCode"))
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedResultMetadata{}, err
		}
	}
	errorMessage, err := normalizeText(errorMessageInput, resultField(fieldPrefix, "errorMessage"))
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedResultMetadata{}, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedResultMetadata{}, err
	}

	return normalizedResultMetadata{
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
	var validation appvalidation.Collector

	if minValue != nil && maxValue != nil && *minValue > *maxValue {
		validation.Add(resultField(fieldPrefix, "rttMinMs"), "must be less than or equal to rttMaxMs", minValue)
	}
	if minValue != nil && avgValue != nil && *minValue > *avgValue {
		validation.Add(resultField(fieldPrefix, "rttMinMs"), "must be less than or equal to rttAvgMs", minValue)
	}
	if avgValue != nil && maxValue != nil && *avgValue > *maxValue {
		validation.Add(resultField(fieldPrefix, "rttAvgMs"), "must be less than or equal to rttMaxMs", avgValue)
	}

	return validation.Err(ErrInvalidInput)
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
