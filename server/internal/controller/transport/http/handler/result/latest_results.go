package result

import (
	"context"
	"time"

	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
)

func (h *Handler) queryLatestResults(ctx context.Context, input *queryLatestResultsInput) (*queryLatestResultsOutput, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.QueryLatestResults(ctx, appresult.QueryLatestResultsInput{
		CurrentUserID: userID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
		CheckID:       input.CheckID,
		Type:          input.Type,
	})
	if err != nil {
		return nil, mapResultError(err, "query latest results failed")
	}

	return &queryLatestResultsOutput{Body: newQueryLatestResultsBody(output)}, nil
}

type queryLatestResultsInput struct {
	Ref     string
	ProbeID string
	CheckID string
	Type    string
}

type queryLatestResultsOutput struct {
	Body queryLatestResultsBody
}

type queryLatestResultsBody struct {
	Results []latestResultBody `json:"results"`
}

type latestResultBody struct {
	Type            string    `json:"type"`
	ProbeID         string    `json:"probeId"`
	CheckID         string    `json:"checkId"`
	LatestStartedAt time.Time `json:"latestStartedAt"`
	LatestStatus    string    `json:"latestStatus"`
}

func newQueryLatestResultsBody(output appresult.LatestResultsOutput) queryLatestResultsBody {
	results := make([]latestResultBody, 0, len(output.Results))
	for _, result := range output.Results {
		results = append(results, latestResultBody{
			Type:            result.Type,
			ProbeID:         result.ProbeID,
			CheckID:         result.CheckID,
			LatestStartedAt: result.LatestStartedAt,
			LatestStatus:    result.LatestStatus,
		})
	}

	return queryLatestResultsBody{Results: results}
}
