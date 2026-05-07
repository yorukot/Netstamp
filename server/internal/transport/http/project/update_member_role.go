package project

import (
	"context"

	appproject "github.com/yorukot/netstamp/internal/application/project"
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

	return &memberOutput{Body: memberOutputBody{Member: newProjectMemberResponse(member)}}, nil
}

type updateMemberRoleInput struct {
	Ref    string `path:"ref" minLength:"1" maxLength:"100" doc:"Project UUID or slug." example:"engineering"`
	UserID string `path:"user_id" format:"uuid"`
	Body   updateMemberRoleInputBody
}

type updateMemberRoleInputBody struct {
	Role string `json:"role" enum:"owner,admin,editor,viewer" required:"true" doc:"Project member role." example:"viewer"`
}
