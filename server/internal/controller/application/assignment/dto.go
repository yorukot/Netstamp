package assignment

import (
	"encoding/json"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
)

type PreviewSelectorInput struct {
	CurrentUserID string
	ProjectRef    string
	Selector      json.RawMessage
}

type ListProjectAssignmentsInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	CheckID       string
}

type SelectorPreviewOutput struct {
	Selector     json.RawMessage
	MatchedCount int32
	Probes       []domainprobe.Probe
}

type ProjectAssignmentsOutput struct {
	Assignments []domainassignment.Assignment
}

type normalizedPreviewSelectorInput struct {
	currentUserID string
	projectRef    string
	selector      domainselector.Selector
}

type normalizedListProjectAssignmentsInput struct {
	currentUserID string
	projectRef    string
	probeID       string
	checkID       string
}
