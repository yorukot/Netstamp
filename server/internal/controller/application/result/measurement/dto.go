package measurement

import (
	"time"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
)

type QueryInput struct {
	CurrentUserID string
	ProjectRef    string
	ProbeID       string
	CheckID       string
	Type          string
	Status        string
	FromMs        *int64
	ToMs          *int64
	Limit         *int32
	CursorMs      *int64
	Now           time.Time
}

type Output struct {
	Measurements []Measurement
	Query        QueryMetadata
}

type Measurement struct {
	Type         string
	StartedAt    time.Time
	FinishedAt   time.Time
	ProbeID      string
	CheckID      string
	Status       string
	DurationMs   int32
	LatencyMs    *float64
	LossPercent  *float64
	Metadata     *string
	ErrorCode    *string
	ErrorMessage *string
}

type QueryMetadata struct {
	FromMs     int64
	ToMs       int64
	Limit      int32
	NextCursor *int64
}

type normalizedInput struct {
	base       resultshared.QueryBase
	resultType *string
	status     *string
	limit      int32
	cursor     *time.Time
}
