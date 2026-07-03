package assignment

import (
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
)

func normalizeProjectTarget(projectID string) (string, error) {
	projectID, err := domainassignment.VNProjectID(projectID)
	if err != nil {
		return "", appvalidation.New(ErrInvalidInput, "projectId", err.Error(), projectID)
	}

	return projectID, nil
}

func normalizeProbeTarget(projectID, probeID string) (string, string, error) {
	var validation appvalidation.Collector

	projectID, err := domainassignment.VNProjectID(projectID)
	if err != nil {
		validation.AddError("projectId", err, projectID)
	}
	probeID, err = domainassignment.VNProbeID(probeID)
	if err != nil {
		validation.AddError("probeId", err, probeID)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return "", "", err
	}

	return projectID, probeID, nil
}

func normalizeCheckTarget(projectID, checkID string) (string, string, error) {
	var validation appvalidation.Collector

	projectID, err := domainassignment.VNProjectID(projectID)
	if err != nil {
		validation.AddError("projectId", err, projectID)
	}
	checkID, err = domainassignment.VNCheckID(checkID)
	if err != nil {
		validation.AddError("checkId", err, checkID)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return "", "", err
	}

	return projectID, checkID, nil
}

func normalizeLabelTarget(projectID, labelID string) (string, string, error) {
	var validation appvalidation.Collector

	projectID, err := domainassignment.VNProjectID(projectID)
	if err != nil {
		validation.AddError("projectId", err, projectID)
	}
	labelID, err = domainassignment.VNLabelID(labelID)
	if err != nil {
		validation.AddError("labelId", err, labelID)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return "", "", err
	}

	return projectID, labelID, nil
}

func normalizePreviewSelectorInput(input PreviewSelectorInput) (normalizedPreviewSelectorInput, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	selector, err := domainselector.Parse(input.Selector)
	if err != nil {
		validation.AddError("selector", err, nil)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedPreviewSelectorInput{}, err
	}

	return normalizedPreviewSelectorInput{
		currentUserID: input.CurrentUserID,
		projectRef:    projectRef,
		selector:      selector,
	}, nil
}

func normalizeListProjectAssignmentsInput(input ListProjectAssignmentsInput) (normalizedListProjectAssignmentsInput, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	var probeID string
	if input.ProbeID != "" {
		probeID, err = domainprobe.VNProbeID(input.ProbeID)
		if err != nil {
			validation.AddError("probeId", err, input.ProbeID)
		}
	}
	var checkID string
	if input.CheckID != "" {
		checkID, err = domaincheck.VNCheckID(input.CheckID)
		if err != nil {
			validation.AddError("checkId", err, input.CheckID)
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedListProjectAssignmentsInput{}, err
	}

	return normalizedListProjectAssignmentsInput{
		currentUserID: input.CurrentUserID,
		projectRef:    projectRef,
		probeID:       probeID,
		checkID:       checkID,
	}, nil
}
