package check

import (
	"context"

	appcheck "github.com/yorukot/netstamp/internal/application/check"
)

func (h *Handler) deleteCheck(ctx context.Context, input *deleteCheckInput) (*struct{}, error) {
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

	return nil, nil
}

type deleteCheckInput struct {
	Ref     string `path:"ref" minLength:"1" maxLength:"100" doc:"Project UUID or slug." example:"engineering"`
	CheckID string `path:"check_id" format:"uuid" doc:"Check ID." example:"33333333-3333-3333-3333-333333333333"`
}
