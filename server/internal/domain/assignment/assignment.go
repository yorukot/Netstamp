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
	ID              string
	ProjectID       string
	ProbeID         string
	CheckID         string
	CheckVersion    string
	SelectorVersion string
	Check           *check.Check
	Probe           *probe.Probe
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
