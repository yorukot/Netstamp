package result

import (
	"strings"
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
	var validation appvalidation.Collector

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
		if !validation.AddValidation(err) {
			return normalizedQueryPingSeriesInput{}, err
		}
	}
	metric, maxDataPoints, err := normalizePingSeriesOptions(input)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedQueryPingSeriesInput{}, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedQueryPingSeriesInput{}, err
	}

	return normalizedQueryPingSeriesInput{
		normalizedQueryBase: base,
		metric:              metric,
		maxDataPoints:       maxDataPoints,
	}, nil
}

func normalizeQueryTracerouteRunsInput(input QueryTracerouteRunsInput) (normalizedQueryTracerouteRunsInput, error) {
	var validation appvalidation.Collector

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
		if !validation.AddValidation(err) {
			return normalizedQueryTracerouteRunsInput{}, err
		}
	}
	limit, cursor, err := normalizeTracerouteRunOptions(input)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedQueryTracerouteRunsInput{}, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedQueryTracerouteRunsInput{}, err
	}

	return normalizedQueryTracerouteRunsInput{
		normalizedQueryBase: base,
		limit:               limit,
		cursor:              cursor,
	}, nil
}

func normalizeQueryTracerouteTopologyInput(input QueryTracerouteTopologyInput) (normalizedQueryTracerouteTopologyInput, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	probeID, err := normalizeOptionalProbeID(input.ProbeID)
	if err != nil {
		validation.AddError("probeId", err, input.ProbeID)
	}
	checkID, err := normalizeOptionalCheckID(input.CheckID)
	if err != nil {
		validation.AddError("checkId", err, input.CheckID)
	}
	now := input.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	from, to, err := normalizeRange(input.FromMs, input.ToMs, now)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedQueryTracerouteTopologyInput{}, err
		}
	}
	limit, err := normalizeRunLimit(input.Limit)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedQueryTracerouteTopologyInput{}, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedQueryTracerouteTopologyInput{}, err
	}

	return normalizedQueryTracerouteTopologyInput{
		currentUserID: input.CurrentUserID,
		projectRef:    projectRef,
		probeID:       probeID,
		checkID:       checkID,
		from:          from,
		to:            to,
		limit:         limit,
	}, nil
}

func normalizeQueryMeasurementsInput(input QueryMeasurementsInput) (normalizedQueryMeasurementsInput, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	probeID, err := normalizeOptionalProbeID(input.ProbeID)
	if err != nil {
		validation.AddError("probeId", err, input.ProbeID)
	}
	checkID, err := normalizeOptionalCheckID(input.CheckID)
	if err != nil {
		validation.AddError("checkId", err, input.CheckID)
	}
	resultType, err := normalizeMeasurementType(input.Type)
	if err != nil {
		validation.AddError("type", err, input.Type)
	}
	status, err := normalizeMeasurementStatus(input.Status)
	if err != nil {
		validation.AddError("status", err, input.Status)
	}

	now := input.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	from, to, err := normalizeRange(input.FromMs, input.ToMs, now)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedQueryMeasurementsInput{}, err
		}
	}
	limit, err := normalizeRunLimit(input.Limit)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedQueryMeasurementsInput{}, err
		}
	}
	cursor, err := normalizeCursor(input.CursorMs)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedQueryMeasurementsInput{}, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedQueryMeasurementsInput{}, err
	}

	return normalizedQueryMeasurementsInput{
		currentUserID: input.CurrentUserID,
		projectRef:    projectRef,
		probeID:       probeID,
		checkID:       checkID,
		resultType:    resultType,
		status:        status,
		from:          from,
		to:            to,
		limit:         limit,
		cursor:        cursor,
	}, nil
}

func normalizePingSeriesOptions(input QueryPingSeriesInput) (PingMetric, int32, error) {
	var validation appvalidation.Collector

	metric, err := normalizeMetric(input.Metric)
	if err != nil {
		if !validation.AddValidation(err) {
			return "", 0, err
		}
	}
	maxDataPoints, err := normalizeMaxDataPoints(input.MaxDataPoints)
	if err != nil {
		if !validation.AddValidation(err) {
			return "", 0, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return "", 0, err
	}

	return metric, maxDataPoints, nil
}

func normalizeTracerouteRunOptions(input QueryTracerouteRunsInput) (int32, *time.Time, error) {
	var validation appvalidation.Collector

	limit, err := normalizeRunLimit(input.Limit)
	if err != nil {
		if !validation.AddValidation(err) {
			return 0, nil, err
		}
	}
	cursor, err := normalizeCursor(input.CursorMs)
	if err != nil {
		if !validation.AddValidation(err) {
			return 0, nil, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return 0, nil, err
	}

	return limit, cursor, nil
}

func normalizeQueryBase(currentUserID, projectRefValue, probeIDValue, checkIDValue string, fromMs, toMs *int64, now time.Time) (normalizedQueryBase, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(projectRefValue)
	if err != nil {
		validation.AddError("projectRef", err, projectRefValue)
	}
	probeID, err := domainprobe.VNProbeID(probeIDValue)
	if err != nil {
		validation.AddError("probeId", err, probeIDValue)
	}
	checkID, err := domaincheck.VNCheckID(checkIDValue)
	if err != nil {
		validation.AddError("checkId", err, checkIDValue)
	}

	now = now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	from, to, err := normalizeRange(fromMs, toMs, now)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedQueryBase{}, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
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

func normalizeOptionalProbeID(probeIDValue string) (string, error) {
	if probeIDValue == "" {
		return "", nil
	}
	return domainprobe.VNProbeID(probeIDValue)
}

func normalizeOptionalCheckID(checkIDValue string) (string, error) {
	if checkIDValue == "" {
		return "", nil
	}
	return domaincheck.VNCheckID(checkIDValue)
}

func normalizeRange(fromMs, toMs *int64, now time.Time) (time.Time, time.Time, error) {
	var validation appvalidation.Collector

	to := now
	if toMs != nil {
		if *toMs <= 0 {
			validation.Add("to", "must be a positive epoch millisecond timestamp", *toMs)
		} else {
			to = time.UnixMilli(*toMs).UTC()
		}
	}

	from := to.Add(-defaultRange)
	if fromMs != nil {
		if *fromMs <= 0 {
			validation.Add("from", "must be a positive epoch millisecond timestamp", *fromMs)
		} else {
			from = time.UnixMilli(*fromMs).UTC()
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return time.Time{}, time.Time{}, err
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

func normalizeMeasurementType(input string) (*string, error) {
	value := strings.TrimSpace(input)
	switch value {
	case "":
		return nil, nil //nolint:nilnil // Nil means no type filter was provided.
	case "ping", "traceroute":
		return &value, nil
	default:
		return nil, invalidResultField("type", "unsupported measurement type", input)
	}
}

func normalizeMeasurementStatus(input string) (*string, error) {
	value := strings.TrimSpace(input)
	switch value {
	case "":
		return nil, nil //nolint:nilnil // Nil means no status filter was provided.
	case "successful", "timeout", "error", "partial":
		return &value, nil
	default:
		return nil, invalidResultField("status", "unsupported measurement status", input)
	}
}

func invalidResultField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
