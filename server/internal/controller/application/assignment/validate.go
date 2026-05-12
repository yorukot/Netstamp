package assignment

import (
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
)

func normalizeProbeTarget(projectID, probeID string) (string, string, error) {
	projectID, err := domainassignment.VNProjectID(projectID)
	if err != nil {
		return "", "", invalidAssignmentField("projectId", err.Error(), projectID)
	}
	probeID, err = domainassignment.VNProbeID(probeID)
	if err != nil {
		return "", "", invalidAssignmentField("probeId", err.Error(), probeID)
	}

	return projectID, probeID, nil
}

func normalizeCheckTarget(projectID, checkID string) (string, string, error) {
	projectID, err := domainassignment.VNProjectID(projectID)
	if err != nil {
		return "", "", invalidAssignmentField("projectId", err.Error(), projectID)
	}
	checkID, err = domainassignment.VNCheckID(checkID)
	if err != nil {
		return "", "", invalidAssignmentField("checkId", err.Error(), checkID)
	}

	return projectID, checkID, nil
}

func normalizeLabelTarget(projectID, labelID string) (string, string, error) {
	projectID, err := domainassignment.VNProjectID(projectID)
	if err != nil {
		return "", "", invalidAssignmentField("projectId", err.Error(), projectID)
	}
	labelID, err = domainassignment.VNLabelID(labelID)
	if err != nil {
		return "", "", invalidAssignmentField("labelId", err.Error(), labelID)
	}

	return projectID, labelID, nil
}

func invalidAssignmentField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
