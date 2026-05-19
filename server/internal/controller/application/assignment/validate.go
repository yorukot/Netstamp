package assignment

import (
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
)

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
