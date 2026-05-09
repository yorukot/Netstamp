package check

import (
	"context"

	appcheck "github.com/yorukot/netstamp/internal/application/check"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

func (h *Handler) listChecks(ctx context.Context, input *listChecksInput) (*listChecksOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	checks, err := h.service.ListChecks(ctx, appcheck.ListChecksInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
	})
	if err != nil {
		return nil, mapCheckError(err, "list checks failed")
	}

	return &listChecksOutput{Body: listChecksOutputBody{Checks: newCheckResponses(checks)}}, nil
}

type listChecksInput struct {
	Ref string `path:"ref" doc:"Project UUID or slug." example:"engineering"`
}

type listChecksOutput struct {
	Body listChecksOutputBody
}

type listChecksOutputBody struct {
	Checks []checkResponse `json:"checks"`
}

func newCheckResponses(checks []domaincheck.Check) []checkResponse {
	responses := make([]checkResponse, 0, len(checks))
	for _, check := range checks {
		responses = append(responses, newCheckResponse(check))
	}

	return responses
}
