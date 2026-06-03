package shared

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

func NormalizeQueryBase(currentUserID, projectRefValue, probeIDValue, checkIDValue string, fromMs, toMs *int64, now time.Time) (QueryBase, error) {
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
	from, to, err := NormalizeRange(fromMs, toMs, now)
	if err != nil {
		if !validation.AddValidation(err) {
			return QueryBase{}, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return QueryBase{}, err
	}

	return QueryBase{
		CurrentUserID: currentUserID,
		ProjectRef:    projectRef,
		ProbeID:       probeID,
		CheckID:       checkID,
		From:          from,
		To:            to,
		Now:           now,
	}, nil
}

func NormalizeOptionalProbeID(probeIDValue string) (string, error) {
	if probeIDValue == "" {
		return "", nil
	}
	return domainprobe.VNProbeID(probeIDValue)
}

func NormalizeOptionalCheckID(checkIDValue string) (string, error) {
	if checkIDValue == "" {
		return "", nil
	}
	return domaincheck.VNCheckID(checkIDValue)
}

func NormalizeRange(fromMs, toMs *int64, now time.Time) (time.Time, time.Time, error) {
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
		return time.Time{}, time.Time{}, InvalidField("from", "must be before to", from.UnixMilli())
	}

	return from, to, nil
}

func NormalizeMaxDataPoints(input *int32) (int32, error) {
	if input == nil {
		return defaultMaxDataPoint, nil
	}
	if *input < 1 {
		return 0, InvalidField("maxDataPoints", "must be greater than or equal to 1", *input)
	}
	if *input > maxDataPointsLimit {
		return 0, InvalidField("maxDataPoints", "must be less than or equal to 2000", *input)
	}

	return *input, nil
}

func NormalizeRunLimit(input *int32) (int32, error) {
	if input == nil {
		return defaultRunLimit, nil
	}
	if *input < 1 {
		return 0, InvalidField("limit", "must be greater than or equal to 1", *input)
	}
	if *input > maxRunLimit {
		return 0, InvalidField("limit", "must be less than or equal to 500", *input)
	}

	return *input, nil
}

func NormalizeCursor(input *int64) (*time.Time, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means no pagination cursor was provided.
	}
	if *input <= 0 {
		return nil, InvalidField("cursor", "must be a positive epoch millisecond timestamp", *input)
	}
	cursor := time.UnixMilli(*input).UTC()
	return &cursor, nil
}

func InvalidField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
