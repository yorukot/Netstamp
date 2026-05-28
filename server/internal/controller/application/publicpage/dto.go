package publicpage

import (
	"time"
)

type ListPagesInput struct {
	CurrentUserID string
	ProjectRef    string
}

type GetPageInput struct {
	CurrentUserID string
	ProjectRef    string
	PageID        string
}

type CreatePageInput struct {
	CurrentUserID string
	ProjectRef    string
	Slug          string
	Title         string
	Description   *string
	Enabled       bool
}

type UpdatePageInput struct {
	CurrentUserID  string
	ProjectRef     string
	PageID         string
	Slug           *string
	Title          *string
	Description    *string
	DescriptionSet bool
	Enabled        *bool
}

type DeletePageInput struct {
	CurrentUserID string
	ProjectRef    string
	PageID        string
}

type CreateFolderInput struct {
	CurrentUserID string
	ProjectRef    string
	PageID        string
	ParentID      *string
	Name          string
	Description   *string
	SortOrder     int32
}

type UpdateFolderInput struct {
	CurrentUserID  string
	ProjectRef     string
	PageID         string
	FolderID       string
	ParentID       *string
	ParentIDSet    bool
	Name           *string
	Description    *string
	DescriptionSet bool
	SortOrder      *int32
}

type DeleteFolderInput struct {
	CurrentUserID string
	ProjectRef    string
	PageID        string
	FolderID      string
}

type SetFolderChecksInput struct {
	CurrentUserID string
	ProjectRef    string
	PageID        string
	FolderID      string
	CheckIDs      []string
}

type GetPublicPageInput struct {
	Slug string
}

type QueryPublicPingInsightInput struct {
	Slug          string
	ProbeID       string
	CheckID       string
	FromMs        *int64
	ToMs          *int64
	MaxDataPoints *int32
	Now           time.Time
}

type PublicPingInsightOutput struct {
	Buckets       []PingInsightBucket
	SampleDensity []PingSampleDensityCell
	Summary       PingInsightSummary
	Query         QueryMetadata
}

type PingInsightBucket struct {
	TimestampMs   int64
	ResultCount   int64
	DurationAvgMs *float64
	RttMinMs      *float64
	RttAvgMs      *float64
	RttMedianMs   *float64
	RttMaxMs      *float64
	RttStddevMs   *float64
	LossPercent   *float64
	SuccessRate   *float64
	SentCount     int64
	ReceivedCount int64
	TimeoutCount  int64
	ErrorCount    int64
}

type PingSampleDensityCell struct {
	TimestampMs      int64
	RttBucketStartMs float64
	RttBucketEndMs   float64
	SampleCount      int64
}

type PingInsightSummary struct {
	TotalResults      int64
	SuccessfulCount   int64
	TimeoutCount      int64
	ErrorCount        int64
	SentCount         int64
	ReceivedCount     int64
	AvgLossPercent    *float64
	AvgRttMs          *float64
	MedianRttMs       *float64
	MaxRttMs          *float64
	P95RttMs          *float64
	P99RttMs          *float64
	LatestStatus      *string
	LatestStartedAtMs *int64
	LatestRttAvgMs    *float64
	LatestLossPercent *float64
}

type QueryMetadata struct {
	FromMs        int64
	ToMs          int64
	MaxDataPoints int32
	Resolution    string
	TotalPoints   int64
}
