package measurement

import (
	"strings"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
)

func normalizeQueryInput(input QueryInput) (normalizedInput, error) {
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
			return normalizedInput{}, err
		}
	}
	resultType, err := normalizeType(input.Type)
	if err != nil {
		validation.AddError("type", err, input.Type)
	}
	status, err := normalizeStatus(input.Status)
	if err != nil {
		validation.AddError("status", err, input.Status)
	}
	limit, err := resultshared.NormalizeRunLimit(input.Limit)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedInput{}, err
		}
	}
	cursor, err := resultshared.NormalizeCursor(input.CursorMs)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedInput{}, err
		}
	}
	if err := validation.Err(resultshared.ErrInvalidInput); err != nil {
		return normalizedInput{}, err
	}

	return normalizedInput{
		base:       base,
		resultType: resultType,
		status:     status,
		limit:      limit,
		cursor:     cursor,
	}, nil
}

func normalizeType(input string) (*string, error) {
	value := strings.TrimSpace(input)
	switch value {
	case "":
		return nil, nil //nolint:nilnil // Nil means no type filter was provided.
	case "ping", "tcp", "traceroute":
		return &value, nil
	default:
		return nil, resultshared.InvalidField("type", "unsupported measurement type", input)
	}
}

func normalizeStatus(input string) (*string, error) {
	value := strings.TrimSpace(input)
	switch value {
	case "":
		return nil, nil //nolint:nilnil // Nil means no status filter was provided.
	case "successful", "timeout", "error", "partial":
		return &value, nil
	default:
		return nil, resultshared.InvalidField("status", "unsupported measurement status", input)
	}
}
