package result

import (
	"time"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

const (
	defaultRange        = 24 * time.Hour
	defaultMaxDataPoint = int32(600)
	maxDataPointsLimit  = int32(2000)
)

func normalizeQueryPingSeriesInput(input QueryPingSeriesInput) (normalizedQueryPingSeriesInput, error) {
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		return normalizedQueryPingSeriesInput{}, invalidResultField("projectRef", err.Error(), input.ProjectRef)
	}
	probeID, err := domainprobe.VNProbeID(input.ProbeID)
	if err != nil {
		return normalizedQueryPingSeriesInput{}, invalidResultField("probeId", err.Error(), input.ProbeID)
	}
	checkID, err := domaincheck.VNCheckID(input.CheckID)
	if err != nil {
		return normalizedQueryPingSeriesInput{}, invalidResultField("checkId", err.Error(), input.CheckID)
	}

	now := input.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	from, to, err := normalizeRange(input.FromMs, input.ToMs, now)
	if err != nil {
		return normalizedQueryPingSeriesInput{}, err
	}
	metric, err := normalizeMetric(input.Metric)
	if err != nil {
		return normalizedQueryPingSeriesInput{}, err
	}
	maxDataPoints, err := normalizeMaxDataPoints(input.MaxDataPoints)
	if err != nil {
		return normalizedQueryPingSeriesInput{}, err
	}

	return normalizedQueryPingSeriesInput{
		currentUserID: input.CurrentUserID,
		projectRef:    projectRef,
		probeID:       probeID,
		checkID:       checkID,
		from:          from,
		to:            to,
		metric:        metric,
		maxDataPoints: maxDataPoints,
	}, nil
}

func normalizeRange(fromMs, toMs *int64, now time.Time) (time.Time, time.Time, error) {
	to := now
	if toMs != nil {
		if *toMs <= 0 {
			return time.Time{}, time.Time{}, invalidResultField("to", "must be a positive epoch millisecond timestamp", *toMs)
		}
		to = time.UnixMilli(*toMs).UTC()
	}

	from := to.Add(-defaultRange)
	if fromMs != nil {
		if *fromMs <= 0 {
			return time.Time{}, time.Time{}, invalidResultField("from", "must be a positive epoch millisecond timestamp", *fromMs)
		}
		from = time.UnixMilli(*fromMs).UTC()
	}
	if !from.Before(to) {
		return time.Time{}, time.Time{}, invalidResultField("from", "must be before to", from.UnixMilli())
	}

	return from, to, nil
}

func normalizeMetric(input string) (PingMetric, error) {
	switch PingMetric(input) {
	case "":
		return PingMetricRTTAvgMS, nil
	case PingMetricRTTAvgMS:
		return PingMetricRTTAvgMS, nil
	case PingMetricLossPercent:
		return PingMetricLossPercent, nil
	case PingMetricSuccessRate:
		return PingMetricSuccessRate, nil
	default:
		return "", invalidResultField("metric", "unsupported ping metric", input)
	}
}

func normalizeMaxDataPoints(input *int32) (int32, error) {
	if input == nil {
		return defaultMaxDataPoint, nil
	}
	if *input < 1 {
		return 0, invalidResultField("maxDataPoints", "must be greater than or equal to 1", *input)
	}
	if *input > maxDataPointsLimit {
		return 0, invalidResultField("maxDataPoints", "must be less than or equal to 2000", *input)
	}

	return *input, nil
}

func invalidResultField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
