package assignment

import (
	"encoding/json"
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type projectAssignmentBody struct {
	ID              string                      `json:"id"`
	ProjectID       string                      `json:"projectId"`
	ProbeID         string                      `json:"probeId"`
	CheckID         string                      `json:"checkId"`
	CheckVersion    string                      `json:"checkVersion"`
	SelectorVersion string                      `json:"selectorVersion"`
	Probe           *domainprobe.Probe          `json:"probe,omitempty"`
	Check           *projectAssignmentCheckBody `json:"check,omitempty"`
}

type projectAssignmentCheckBody struct {
	ID               string                   `json:"id"`
	ProjectID        string                   `json:"projectId"`
	Name             string                   `json:"name"`
	Type             domaincheck.Type         `json:"type"`
	Target           string                   `json:"target"`
	Selector         json.RawMessage          `json:"selector,omitempty"`
	Description      *string                  `json:"description,omitempty"`
	IntervalSeconds  int32                    `json:"intervalSeconds"`
	Labels           []domainlabel.Label      `json:"labels"`
	CreatedAt        time.Time                `json:"createdAt"`
	UpdatedAt        time.Time                `json:"updatedAt"`
	PingConfig       *domainping.Config       `json:"pingConfig,omitempty"`
	TCPConfig        *domaintcp.Config        `json:"tcpConfig,omitempty"`
	TracerouteConfig *domaintraceroute.Config `json:"tracerouteConfig,omitempty"`
}

func newProjectAssignmentBodies(assignments []domainassignment.Assignment) []projectAssignmentBody {
	values := make([]projectAssignmentBody, 0, len(assignments))
	for _, assignment := range assignments {
		values = append(values, newProjectAssignmentBody(assignment))
	}
	return values
}

func newProjectAssignmentBody(assignment domainassignment.Assignment) projectAssignmentBody {
	return projectAssignmentBody{
		ID:              assignment.ID,
		ProjectID:       assignment.ProjectID,
		ProbeID:         assignment.ProbeID,
		CheckID:         assignment.CheckID,
		CheckVersion:    assignment.CheckVersion,
		SelectorVersion: assignment.SelectorVersion,
		Probe:           assignment.Probe,
		Check:           newProjectAssignmentCheckBody(assignment.Check),
	}
}

func newProjectAssignmentCheckBody(check *domaincheck.Check) *projectAssignmentCheckBody {
	if check == nil {
		return nil
	}
	target := check.Target
	if check.Type == domaincheck.TypeHTTP {
		target = domainhttp.RedactTarget(target)
	}
	return &projectAssignmentCheckBody{
		ID:               check.ID,
		ProjectID:        check.ProjectID,
		Name:             check.Name,
		Type:             check.Type,
		Target:           target,
		Selector:         append(json.RawMessage(nil), check.Selector...),
		Description:      check.Description,
		IntervalSeconds:  check.IntervalSeconds,
		Labels:           check.Labels,
		CreatedAt:        check.CreatedAt,
		UpdatedAt:        check.UpdatedAt,
		PingConfig:       check.PingConfig,
		TCPConfig:        check.TCPConfig,
		TracerouteConfig: check.TracerouteConfig,
	}
}
