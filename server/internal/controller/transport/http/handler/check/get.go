package check

import (
	"context"

	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
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

	return &checkOutput{Body: checkOutputBody{Check: check}}, nil
}

type getCheckInput struct {
	Ref     string
	CheckID string
}
