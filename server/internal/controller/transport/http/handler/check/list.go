package check

import (
	"context"

	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
)

func (h *Handler) listChecks(ctx context.Context, input *listChecksInput) (*listChecksOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.ListChecks(ctx, appcheck.ListChecksInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
	})
	if err != nil {
		return nil, mapCheckError(err, "list checks failed")
	}

	checks := make([]checkBody, 0, len(output.Checks))
	for _, check := range output.Checks {
		checks = append(checks, newCheckBody(check, output.CanManageChecks))
	}

	return &listChecksOutput{Body: listChecksOutputBody{
		Checks:          checks,
		CanManageChecks: output.CanManageChecks,
	}}, nil
}

type listChecksInput struct {
	Ref string
}

type listChecksOutput struct {
	Body listChecksOutputBody
}

type listChecksOutputBody struct {
	Checks          []checkBody `json:"checks"`
	CanManageChecks bool        `json:"canManageChecks"`
}
