package traceroute

import (
	"time"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func normalizeQueryRunsInput(input QueryRunsInput) (normalizedRunsInput, error) {
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
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedRunsInput{}, err
		}
	}
	limit, cursor, err := normalizeRunOptions(input)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedRunsInput{}, err
		}
	}
	if err := validation.Err(resultshared.ErrInvalidInput); err != nil {
		return normalizedRunsInput{}, err
	}

	return normalizedRunsInput{
		base:   base,
		limit:  limit,
		cursor: cursor,
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
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedInsightInput{}, err
		}
	}
	maxDataPoints, err := resultshared.NormalizeMaxDataPoints(input.MaxDataPoints)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedInsightInput{}, err
		}
	}
	if err := validation.Err(resultshared.ErrInvalidInput); err != nil {
		return normalizedInsightInput{}, err
	}

	return normalizedInsightInput{
		base:          base,
		maxDataPoints: maxDataPoints,
	}, nil
}

func normalizeQueryTopologyInput(input QueryTopologyInput) (normalizedTopologyInput, error) {
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
	now := input.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	from, to, err := resultshared.NormalizeRange(input.FromMs, input.ToMs, now)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedTopologyInput{}, err
		}
	}
	limit, err := resultshared.NormalizeRunLimit(input.Limit)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedTopologyInput{}, err
		}
	}
	if err := validation.Err(resultshared.ErrInvalidInput); err != nil {
		return normalizedTopologyInput{}, err
	}

	return normalizedTopologyInput{
		currentUserID: input.CurrentUserID,
		projectRef:    projectRef,
		probeID:       probeID,
		checkID:       checkID,
		from:          from,
		to:            to,
		limit:         limit,
	}, nil
}

func normalizeRunOptions(input QueryRunsInput) (int32, *time.Time, error) {
	var validation appvalidation.Collector

	limit, err := resultshared.NormalizeRunLimit(input.Limit)
	if err != nil {
		if !validation.AddValidation(err) {
			return 0, nil, err
		}
	}
	cursor, err := resultshared.NormalizeCursor(input.CursorMs)
	if err != nil {
		if !validation.AddValidation(err) {
			return 0, nil, err
		}
	}
	if err := validation.Err(resultshared.ErrInvalidInput); err != nil {
		return 0, nil, err
	}

	return limit, cursor, nil
}
