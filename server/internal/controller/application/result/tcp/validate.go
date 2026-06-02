package tcp

import (
	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
)

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
