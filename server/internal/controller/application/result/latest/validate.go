package latest

import (
	"strings"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func normalizeQueryInput(input QueryInput) (normalizedInput, error) {
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
	resultType, err := normalizeType(input.Type)
	if err != nil {
		validation.AddError("type", err, input.Type)
	}
	if err := validation.Err(resultshared.ErrInvalidInput); err != nil {
		return normalizedInput{}, err
	}

	return normalizedInput{
		currentUserID: input.CurrentUserID,
		projectRef:    projectRef,
		probeID:       probeID,
		checkID:       checkID,
		resultType:    resultType,
	}, nil
}

func normalizeType(input string) (*string, error) {
	value := strings.TrimSpace(input)
	if value == "" {
		return nil, nil //nolint:nilnil // Nil means no result type filter was provided.
	}

	resultType, err := domaincheck.VNCheckType(domaincheck.Type(value))
	if err != nil {
		return nil, resultshared.InvalidField("type", "unsupported result type", input)
	}
	normalized := string(resultType)
	return &normalized, nil
}
