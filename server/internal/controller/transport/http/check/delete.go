package check

import (
	"context"

	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
)

func (h *Handler) deleteCheck(ctx context.Context, input *deleteCheckInput) (*deleteCheckOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.DeleteCheck(ctx, appcheck.GetCheckInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		CheckID:       input.CheckID,
	}); err != nil {
		return nil, mapCheckError(err, "delete check failed")
	}

	return &deleteCheckOutput{}, nil
}

type deleteCheckInput struct {
	Ref     string `path:"ref" minLength:"1" maxLength:"64" pattern:"^[a-z0-9-]+$" patternDescription:"lowercase letters, numbers, and dashes" doc:"Project slug or lowercase UUID." example:"engineering"`
	CheckID string `path:"check_id" minLength:"1" format:"uuid" doc:"Check ID." example:"33333333-3333-3333-3333-333333333333"`
}

type deleteCheckOutput struct{}
