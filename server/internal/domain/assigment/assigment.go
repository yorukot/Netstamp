package assigment

import (
	"github.com/yorukot/netstamp/internal/domain/check"
	"github.com/yorukot/netstamp/internal/domain/probe"
)

type Assigment struct {
	ID              string       `json:"id"`
	ProbeID         string       `json:"probe_id"`
	CheckID         string       `json:"check_id"`
	CheckVersion    string       `json:"check_version"`
	SelectorVersion string       `json:"probe_version"`
	Check           *check.Check `json:"check"`
	Probe           *probe.Probe `json:"probe"`
}
