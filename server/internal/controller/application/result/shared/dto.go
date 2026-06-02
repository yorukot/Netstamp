package shared

import "time"

type QueryBase struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	CheckID       string
	From          time.Time
	To            time.Time
}

type QueryMetadata struct {
	FromMs        int64
	ToMs          int64
	MaxDataPoints int32
	Source        string
	Resolution    string
	TotalPoints   int64
}
