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
	Summary PingInsightSummary
	Meta    QueryMetadata
}

type PingInsightSummary struct {
	AverageRttMs *float64
	MaxRttMs     *float64
	LossPercent  *float64
	SuccessRate  *float64
	Samples      int64
}

type QueryMetadata struct {
	FromMs        int64
	ToMs          int64
	MaxDataPoints int32
	Source        string
	Resolution    string
	TotalPoints   int64
}
