package project

import (
	"context"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func (h *Handler) updateMemberRole(ctx context.Context, input *updateMemberRoleInput) (*memberOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	member, err := h.service.UpdateMemberRole(ctx, appproject.UpdateMemberRoleInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		UserID:        input.UserID,
		Role:          domainproject.Role(input.Body.Role),
	})
	if err != nil {
		return nil, mapProjectError(err, "update project member failed")
	}

	return &memberOutput{Body: memberOutputBody{Member: member}}, nil
}

type updateMemberRoleInput struct {
	Ref    string
	UserID string
	Body   updateMemberRoleInputBody
}

type updateMemberRoleInputBody struct {
	Role string `json:"role,omitempty"`
}
