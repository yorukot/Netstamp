package check

import (
	"context"

	appcheck "github.com/yorukot/netstamp/internal/application/check"
)

func (h *Handler) getCheck(ctx context.Context, input *getCheckInput) (*checkOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	check, err := h.service.GetCheck(ctx, appcheck.GetCheckInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		CheckID:       input.CheckID,
	})
	if err != nil {
		return nil, mapCheckError(err, "get check failed")
	}

	return &checkOutput{Body: checkOutputBody{Check: newCheckResponse(check)}}, nil
}

type getCheckInput struct {
	Ref     string `path:"ref" minLength:"1" maxLength:"100" doc:"Project UUID or slug." example:"engineering"`
	CheckID string `path:"check_id" format:"uuid" doc:"Check ID." example:"33333333-3333-3333-3333-333333333333"`
}
