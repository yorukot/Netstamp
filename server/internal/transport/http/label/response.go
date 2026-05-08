package label

import (
	"time"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
)

type labelOutput struct {
	Body labelOutputBody
}

type labelOutputBody struct {
	Label labelResponse `json:"label"`
}

type labelResponse struct {
	ID        string    `json:"id" format:"uuid"`
	ProjectID string    `json:"projectId" format:"uuid"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func newLabelResponse(label domainlabel.Label) labelResponse {
	return labelResponse{
		ID:        label.ID,
		ProjectID: label.ProjectID,
		Key:       label.Key,
		Value:     label.Value,
		CreatedAt: label.CreatedAt,
		UpdatedAt: label.UpdatedAt,
	}
}
