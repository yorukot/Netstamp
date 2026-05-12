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
	Ref     string `path:"ref" minLength:"1" maxLength:"64" pattern:"^[a-z0-9-]+$" patternDescription:"lowercase letters, numbers, and dashes" doc:"Project slug or lowercase UUID." example:"engineering"`
	CheckID string `path:"check_id" minLength:"1" format:"uuid" doc:"Check ID." example:"33333333-3333-3333-3333-333333333333"`
}
