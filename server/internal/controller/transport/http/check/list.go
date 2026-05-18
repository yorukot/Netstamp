package check

import (
	"context"

	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
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

	return &listChecksOutput{Body: listChecksOutputBody{Checks: checks}}, nil
}

type listChecksInput struct {
	Ref string
}

type listChecksOutput struct {
	Body listChecksOutputBody
}

type listChecksOutputBody struct {
	Checks []domaincheck.Check `json:"checks"`
}
