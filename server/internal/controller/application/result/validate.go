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
	defaultRunLimit     = int32(100)
	maxRunLimit         = int32(500)
)

func normalizeQueryPingSeriesInput(input QueryPingSeriesInput) (normalizedQueryPingSeriesInput, error) {
	base, err := normalizeQueryBase(
		input.CurrentUserID,
		input.ProjectRef,
		input.ProbeID,
		input.CheckID,
		input.FromMs,
		input.ToMs,
		input.Now,
	)
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
		normalizedQueryBase: base,
		metric:              metric,
		maxDataPoints:       maxDataPoints,
	}, nil
}

func normalizeQueryTracerouteRunsInput(input QueryTracerouteRunsInput) (normalizedQueryTracerouteRunsInput, error) {
	base, err := normalizeQueryBase(
		input.CurrentUserID,
		input.ProjectRef,
		input.ProbeID,
		input.CheckID,
		input.FromMs,
		input.ToMs,
		input.Now,
	)
	if err != nil {
		return normalizedQueryTracerouteRunsInput{}, err
	}
	limit, err := normalizeRunLimit(input.Limit)
	if err != nil {
		return normalizedQueryTracerouteRunsInput{}, err
	}
	cursor, err := normalizeCursor(input.CursorMs)
	if err != nil {
		return normalizedQueryTracerouteRunsInput{}, err
	}

	return normalizedQueryTracerouteRunsInput{
		normalizedQueryBase: base,
		limit:               limit,
		cursor:              cursor,
	}, nil
}

func normalizeQueryBase(currentUserID, projectRefValue, probeIDValue, checkIDValue string, fromMs, toMs *int64, now time.Time) (normalizedQueryBase, error) {
	projectRef, err := domainproject.VNProjectRef(projectRefValue)
	if err != nil {
		return normalizedQueryBase{}, invalidResultField("projectRef", err.Error(), projectRefValue)
	}
	probeID, err := domainprobe.VNProbeID(probeIDValue)
	if err != nil {
		return normalizedQueryBase{}, invalidResultField("probeId", err.Error(), probeIDValue)
	}
	checkID, err := domaincheck.VNCheckID(checkIDValue)
	if err != nil {
		return normalizedQueryBase{}, invalidResultField("checkId", err.Error(), checkIDValue)
	}

	now = now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	from, to, err := normalizeRange(fromMs, toMs, now)
	if err != nil {
		return normalizedQueryBase{}, err
	}

	return normalizedQueryBase{
		currentUserID: currentUserID,
		projectRef:    projectRef,
		probeID:       probeID,
		checkID:       checkID,
		from:          from,
		to:            to,
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

func normalizeRunLimit(input *int32) (int32, error) {
	if input == nil {
		return defaultRunLimit, nil
	}
	if *input < 1 {
		return 0, invalidResultField("limit", "must be greater than or equal to 1", *input)
	}
	if *input > maxRunLimit {
		return 0, invalidResultField("limit", "must be less than or equal to 500", *input)
	}

	return *input, nil
}

func normalizeCursor(input *int64) (*time.Time, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means no pagination cursor was provided.
	}
	if *input <= 0 {
		return nil, invalidResultField("cursor", "must be a positive epoch millisecond timestamp", *input)
	}
	cursor := time.UnixMilli(*input).UTC()
	return &cursor, nil
}

func invalidResultField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
