package httpcheck

import (
	"strings"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func normalizeQueryLatestInput(input QueryLatestInput) (normalizedLatestInput, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	probeID, err := resultshared.NormalizeOptionalProbeID(input.ProbeID)
	if err != nil {
		validation.AddError("probeId", err, input.ProbeID)
	}
	checkID, err := resultshared.NormalizeOptionalCheckID(input.CheckID)
	if err != nil {
		validation.AddError("checkId", err, input.CheckID)
	}
	if err := validation.Err(resultshared.ErrInvalidInput); err != nil {
		return normalizedLatestInput{}, err
	}

	return normalizedLatestInput{
		currentUserID: input.CurrentUserID,
		projectRef:    projectRef,
		probeID:       probeID,
		checkID:       checkID,
	}, nil
}

func normalizeQuerySeriesInput(input QuerySeriesInput) (normalizedSeriesInput, error) {
	var validation appvalidation.Collector

	base, err := resultshared.NormalizeQueryBase(
		input.CurrentUserID,
		input.ProjectRef,
		input.ProbeID,
		input.CheckID,
		input.FromMs,
		input.ToMs,
		input.Now,
	)
	if err != nil && !validation.AddValidation(err) {
		return normalizedSeriesInput{}, err
	}
	series, maxDataPoints, err := normalizeSeriesOptions(input)
	if err != nil && !validation.AddValidation(err) {
		return normalizedSeriesInput{}, err
	}
	if err := validation.Err(resultshared.ErrInvalidInput); err != nil {
		return normalizedSeriesInput{}, err
	}

	return normalizedSeriesInput{
		base:          base,
		series:        series,
		maxDataPoints: maxDataPoints,
	}, nil
}

func normalizeQueryInsightInput(input QueryInsightInput) (normalizedInsightInput, error) {
	var validation appvalidation.Collector

	base, err := resultshared.NormalizeQueryBase(
		input.CurrentUserID,
		input.ProjectRef,
		input.ProbeID,
		input.CheckID,
		input.FromMs,
		input.ToMs,
		input.Now,
	)
	if err != nil && !validation.AddValidation(err) {
		return normalizedInsightInput{}, err
	}
	maxDataPoints, err := resultshared.NormalizeMaxDataPoints(input.MaxDataPoints)
	if err != nil && !validation.AddValidation(err) {
		return normalizedInsightInput{}, err
	}
	if err := validation.Err(resultshared.ErrInvalidInput); err != nil {
		return normalizedInsightInput{}, err
	}

	return normalizedInsightInput{base: base, maxDataPoints: maxDataPoints}, nil
}

func normalizeSeriesOptions(input QuerySeriesInput) ([]SeriesKey, int32, error) {
	var validation appvalidation.Collector

	series, err := normalizeSeries(input.Series)
	if err != nil && !validation.AddValidation(err) {
		return nil, 0, err
	}
	maxDataPoints, err := resultshared.NormalizeMaxDataPoints(input.MaxDataPoints)
	if err != nil && !validation.AddValidation(err) {
		return nil, 0, err
	}
	if err := validation.Err(resultshared.ErrInvalidInput); err != nil {
		return nil, 0, err
	}

	return series, maxDataPoints, nil
}

func normalizeSeries(input string) ([]SeriesKey, error) {
	if strings.TrimSpace(input) == "" {
		return []SeriesKey{
			SeriesDNSAvg,
			SeriesConnectAvg,
			SeriesTLSAvg,
			SeriesTTFBAvg,
			SeriesTotalAvg,
			SeriesFailurePercent,
		}, nil
	}

	parts := strings.Split(input, ",")
	seen := make(map[SeriesKey]struct{}, len(parts))
	series := make([]SeriesKey, 0, len(parts))
	for _, part := range parts {
		key, err := normalizeSeriesKey(part)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		series = append(series, key)
	}
	if len(series) == 0 {
		return nil, resultshared.InvalidField("series", "must include at least one http series", input)
	}

	return series, nil
}

func normalizeSeriesKey(input string) (SeriesKey, error) {
	trimmed := strings.TrimSpace(input)
	switch SeriesKey(trimmed) {
	case SeriesDNSAvg:
		return SeriesDNSAvg, nil
	case SeriesConnectAvg:
		return SeriesConnectAvg, nil
	case SeriesTLSAvg:
		return SeriesTLSAvg, nil
	case SeriesTTFBAvg:
		return SeriesTTFBAvg, nil
	case SeriesTotalAvg:
		return SeriesTotalAvg, nil
	case SeriesFailurePercent:
		return SeriesFailurePercent, nil
	default:
		return "", resultshared.InvalidField("series", "unsupported http series", input)
	}
}
