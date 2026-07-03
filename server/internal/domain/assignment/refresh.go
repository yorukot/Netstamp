package assignment

import (
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
)

type ProbeAssignmentCandidate struct {
	ProbeID   string
	ProjectID string
	Name      string
	Enabled   bool
	Labels    []domainlabel.Label
}

type CheckAssignmentCandidate struct {
	Check           domaincheck.Check
	Selector        domainselector.Selector
	SelectorVersion string
	CheckVersion    string
}

type AssignmentWrite struct {
	ProjectID       string
	ProbeID         string
	CheckID         string
	CheckVersion    string
	SelectorVersion string
}
