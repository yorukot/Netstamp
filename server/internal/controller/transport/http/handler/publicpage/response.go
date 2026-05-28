package publicpage

import (
	"time"

	apppublicpage "github.com/yorukot/netstamp/internal/controller/application/publicpage"
	domainpublicpage "github.com/yorukot/netstamp/internal/domain/publicpage"
)

type pageOutput struct {
	Body pageOutputBody
}

type pageOutputBody struct {
	PublicPage publicPageBody `json:"publicPage"`
}

type pageListOutput struct {
	Body pageListOutputBody
}

type pageListOutputBody struct {
	PublicPages []publicPageBody `json:"publicPages"`
}

type folderOutput struct {
	Body folderOutputBody
}

type folderOutputBody struct {
	Folder folderBody `json:"folder"`
}

type checksOutput struct {
	Body checksOutputBody
}

type checksOutputBody struct {
	Checks []publishedCheckBody `json:"checks"`
}

type pingInsightOutput struct {
	Body pingInsightBody
}

type publicPageBody struct {
	ID          string         `json:"id"`
	Slug        string         `json:"slug"`
	Title       string         `json:"title"`
	Description *string        `json:"description,omitempty"`
	Enabled     bool           `json:"enabled"`
	Folders     []folderBody   `json:"folders,omitempty"`
	Pairs       []pingPairBody `json:"pairs,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

type folderBody struct {
	ID          string               `json:"id"`
	ParentID    *string              `json:"parentId,omitempty"`
	Name        string               `json:"name"`
	Description *string              `json:"description,omitempty"`
	SortOrder   int32                `json:"sortOrder"`
	Checks      []publishedCheckBody `json:"checks,omitempty"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
}

type publishedCheckBody struct {
	ID              string    `json:"id"`
	FolderID        string    `json:"folderId"`
	Name            string    `json:"name"`
	Description     *string   `json:"description,omitempty"`
	IntervalSeconds int32     `json:"intervalSeconds"`
	SortOrder       int32     `json:"sortOrder"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type pingPairBody struct {
	FolderID             string  `json:"folderId"`
	ProbeID              string  `json:"probeId"`
	ProbeName            string  `json:"probeName"`
	ProbeLocationName    *string `json:"probeLocationName,omitempty"`
	ProbeStatus          string  `json:"probeStatus"`
	CheckID              string  `json:"checkId"`
	CheckName            string  `json:"checkName"`
	CheckDescription     *string `json:"checkDescription,omitempty"`
	CheckIntervalSeconds int32   `json:"checkIntervalSeconds"`
}

type pingInsightBody struct {
	Buckets       []pingInsightBucketBody     `json:"buckets"`
	SampleDensity []pingSampleDensityCellBody `json:"sampleDensity"`
	Summary       pingInsightSummaryBody      `json:"summary"`
	Query         queryMetadataBody           `json:"query"`
}

type pingInsightBucketBody struct {
	TimestampMs   int64    `json:"timestampMs"`
	ResultCount   int64    `json:"resultCount"`
	DurationAvgMs *float64 `json:"durationAvgMs,omitempty"`
	RttMinMs      *float64 `json:"rttMinMs,omitempty"`
	RttAvgMs      *float64 `json:"rttAvgMs,omitempty"`
	RttMedianMs   *float64 `json:"rttMedianMs,omitempty"`
	RttMaxMs      *float64 `json:"rttMaxMs,omitempty"`
	RttStddevMs   *float64 `json:"rttStddevMs,omitempty"`
	LossPercent   *float64 `json:"lossPercent,omitempty"`
	SuccessRate   *float64 `json:"successRate,omitempty"`
	SentCount     int64    `json:"sentCount"`
	ReceivedCount int64    `json:"receivedCount"`
	TimeoutCount  int64    `json:"timeoutCount"`
	ErrorCount    int64    `json:"errorCount"`
}

type pingSampleDensityCellBody struct {
	TimestampMs      int64   `json:"timestampMs"`
	RttBucketStartMs float64 `json:"rttBucketStartMs"`
	RttBucketEndMs   float64 `json:"rttBucketEndMs"`
	SampleCount      int64   `json:"sampleCount"`
}

type pingInsightSummaryBody struct {
	TotalResults      int64    `json:"totalResults"`
	SuccessfulCount   int64    `json:"successfulCount"`
	TimeoutCount      int64    `json:"timeoutCount"`
	ErrorCount        int64    `json:"errorCount"`
	SentCount         int64    `json:"sentCount"`
	ReceivedCount     int64    `json:"receivedCount"`
	AvgLossPercent    *float64 `json:"avgLossPercent,omitempty"`
	AvgRttMs          *float64 `json:"avgRttMs,omitempty"`
	MedianRttMs       *float64 `json:"medianRttMs,omitempty"`
	MaxRttMs          *float64 `json:"maxRttMs,omitempty"`
	P95RttMs          *float64 `json:"p95RttMs,omitempty"`
	P99RttMs          *float64 `json:"p99RttMs,omitempty"`
	LatestStatus      *string  `json:"latestStatus,omitempty"`
	LatestStartedAtMs *int64   `json:"latestStartedAtMs,omitempty"`
	LatestRttAvgMs    *float64 `json:"latestRttAvgMs,omitempty"`
	LatestLossPercent *float64 `json:"latestLossPercent,omitempty"`
}

type queryMetadataBody struct {
	FromMs        int64  `json:"from"`
	ToMs          int64  `json:"to"`
	MaxDataPoints int32  `json:"maxDataPoints"`
	Resolution    string `json:"resolution"`
	TotalPoints   int64  `json:"totalPoints"`
}

func newPageBody(page domainpublicpage.Page) publicPageBody {
	return publicPageBody{
		ID:          page.ID,
		Slug:        page.Slug,
		Title:       page.Title,
		Description: page.Description,
		Enabled:     page.Enabled,
		Folders:     newFolderBodies(page.Folders),
		Pairs:       newPingPairBodies(page.Pairs),
		CreatedAt:   page.CreatedAt,
		UpdatedAt:   page.UpdatedAt,
	}
}

func newPageBodies(pages []domainpublicpage.Page) []publicPageBody {
	values := make([]publicPageBody, 0, len(pages))
	for _, page := range pages {
		values = append(values, newPageBody(page))
	}
	return values
}

func newFolderBody(folder domainpublicpage.Folder) folderBody {
	return folderBody{
		ID:          folder.ID,
		ParentID:    folder.ParentID,
		Name:        folder.Name,
		Description: folder.Description,
		SortOrder:   folder.SortOrder,
		Checks:      newPublishedCheckBodies(folder.Checks),
		CreatedAt:   folder.CreatedAt,
		UpdatedAt:   folder.UpdatedAt,
	}
}

func newFolderBodies(folders []domainpublicpage.Folder) []folderBody {
	values := make([]folderBody, 0, len(folders))
	for _, folder := range folders {
		values = append(values, newFolderBody(folder))
	}
	return values
}

func newPublishedCheckBodies(checks []domainpublicpage.PublishedCheck) []publishedCheckBody {
	values := make([]publishedCheckBody, 0, len(checks))
	for _, check := range checks {
		values = append(values, publishedCheckBody{
			ID:              check.ID,
			FolderID:        check.FolderID,
			Name:            check.Name,
			Description:     check.Description,
			IntervalSeconds: check.IntervalSeconds,
			SortOrder:       check.SortOrder,
			CreatedAt:       check.CreatedAt,
			UpdatedAt:       check.UpdatedAt,
		})
	}
	return values
}

func newPingPairBodies(pairs []domainpublicpage.PingPair) []pingPairBody {
	values := make([]pingPairBody, 0, len(pairs))
	for _, pair := range pairs {
		values = append(values, pingPairBody{
			FolderID:             pair.FolderID,
			ProbeID:              pair.ProbeID,
			ProbeName:            pair.ProbeName,
			ProbeLocationName:    pair.ProbeLocationName,
			ProbeStatus:          pair.ProbeStatus,
			CheckID:              pair.CheckID,
			CheckName:            pair.CheckName,
			CheckDescription:     pair.CheckDescription,
			CheckIntervalSeconds: pair.CheckIntervalSeconds,
		})
	}
	return values
}

func newPingInsightBody(output apppublicpage.PublicPingInsightOutput) pingInsightBody {
	return pingInsightBody{
		Buckets:       newPingInsightBucketBodies(output.Buckets),
		SampleDensity: newPingSampleDensityBodies(output.SampleDensity),
		Summary:       newPingInsightSummaryBody(output.Summary),
		Query: queryMetadataBody{
			FromMs:        output.Query.FromMs,
			ToMs:          output.Query.ToMs,
			MaxDataPoints: output.Query.MaxDataPoints,
			Resolution:    output.Query.Resolution,
			TotalPoints:   output.Query.TotalPoints,
		},
	}
}

func newPingInsightBucketBodies(buckets []apppublicpage.PingInsightBucket) []pingInsightBucketBody {
	values := make([]pingInsightBucketBody, 0, len(buckets))
	for _, bucket := range buckets {
		values = append(values, pingInsightBucketBody(bucket))
	}
	return values
}

func newPingSampleDensityBodies(cells []apppublicpage.PingSampleDensityCell) []pingSampleDensityCellBody {
	values := make([]pingSampleDensityCellBody, 0, len(cells))
	for _, cell := range cells {
		values = append(values, pingSampleDensityCellBody(cell))
	}
	return values
}

func newPingInsightSummaryBody(summary apppublicpage.PingInsightSummary) pingInsightSummaryBody {
	return pingInsightSummaryBody(summary)
}
