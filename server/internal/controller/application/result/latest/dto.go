package latest

import "time"

type QueryInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	CheckID       string
	Type          string
}

type Output struct {
	Results []Result
}

type Result struct {
	Type            string
	ProbeID         string
	CheckID         string
	LatestStartedAt time.Time
	LatestStatus    string
}

type normalizedInput struct {
	currentUserID string
	projectRef    string
	probeID       string
	checkID       string
	resultType    *string
}
