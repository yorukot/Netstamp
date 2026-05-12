package assignment

import (
	"errors"
	"strings"

	"github.com/yorukot/netstamp/internal/domain/check"
	"github.com/yorukot/netstamp/internal/domain/probe"
	"github.com/yorukot/spvalidator"
)

var ErrInvalidInput = errors.New("assignment input invalid")

type Assignment struct {
	ID              string       `json:"id"`
	ProjectID       string       `json:"projectId"`
	ProbeID         string       `json:"probeId"`
	CheckID         string       `json:"checkId"`
	CheckVersion    string       `json:"checkVersion"`
	SelectorVersion string       `json:"selectorVersion"`
	Check           *check.Check `json:"check,omitempty"`
	Probe           *probe.Probe `json:"probe,omitempty"`
}

func VNProjectID(projectID string) (string, error) {
	return validUUID(projectID)
}

func VNProbeID(probeID string) (string, error) {
	return validUUID(probeID)
}

func VNCheckID(checkID string) (string, error) {
	return validUUID(checkID)
}

func VNLabelID(labelID string) (string, error) {
	return validUUID(labelID)
}

func validUUID(value string) (string, error) {
	value = strings.TrimSpace(value)

	err := spvalidator.Required(value)
	if err != nil {
		return "", err
	}
	err = spvalidator.UUID(value)
	if err != nil {
		return "", err
	}

	return value, nil
}
