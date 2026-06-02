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
	Summary pingInsightSummaryBody `json:"summary"`
	Meta    queryMetadataBody      `json:"meta"`
}

type pingInsightSummaryBody struct {
	AverageRttMs *float64 `json:"averageRttMs,omitempty"`
	MaxRttMs     *float64 `json:"maxRttMs,omitempty"`
	LossPercent  *float64 `json:"lossPercent,omitempty"`
	SuccessRate  *float64 `json:"successRate,omitempty"`
	Samples      int64    `json:"samples"`
}

type queryMetadataBody struct {
	FromMs        int64  `json:"from"`
	ToMs          int64  `json:"to"`
	MaxDataPoints int32  `json:"maxDataPoints"`
	Source        string `json:"source,omitempty"`
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
		Summary: newPingInsightSummaryBody(output.Summary),
		Meta: queryMetadataBody{
			FromMs:        output.Meta.FromMs,
			ToMs:          output.Meta.ToMs,
			MaxDataPoints: output.Meta.MaxDataPoints,
			Source:        output.Meta.Source,
			Resolution:    output.Meta.Resolution,
			TotalPoints:   output.Meta.TotalPoints,
		},
	}
}

func newPingInsightSummaryBody(summary apppublicpage.PingInsightSummary) pingInsightSummaryBody {
	return pingInsightSummaryBody(summary)
}
