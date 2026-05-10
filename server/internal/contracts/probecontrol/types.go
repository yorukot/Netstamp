package probecontrol

import "time"

type CheckType string

const (
	CheckTypePing CheckType = "ping"
	CheckTypeTCP  CheckType = "tcp"
	CheckTypeHTTP CheckType = "http"
	CheckTypeDNS  CheckType = "dns"
)

type Assignment struct {
	ID              string
	CheckID         string
	CheckVersion    string
	SelectorVersion string
	Type            CheckType
}

type AssignmentSet struct {
	ProbeID     string
	GeneratedAt time.Time
	Assignments []Assignment
}

type ResultBatch struct {
	ProbeID     string
	CollectedAt time.Time
	Results     []Result
}

type Result struct {
	AssignmentID string
	CheckID      string
	CheckVersion string
	Type         CheckType
	Status       string
	StartedAt    time.Time
	FinishedAt   time.Time
}
