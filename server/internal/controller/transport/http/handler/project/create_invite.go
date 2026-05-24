package project

import (
	"context"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func (h *Handler) createInvite(ctx context.Context, input *createInviteInput) (*inviteOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	invite, err := h.service.CreateInvite(ctx, appproject.CreateInviteInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		Email:         input.Body.Email,
		Role:          domainproject.Role(input.Body.Role),
	})
	if err != nil {
		return nil, mapProjectError(err, "create project invite failed")
	}

	return &inviteOutput{Body: inviteOutputBody{Invite: invite}}, nil
}

type createInviteInput struct {
	Ref  string
	Body createInviteInputBody
}

type createInviteInputBody struct {
	Email string `json:"email,omitempty"`
	Role  string `json:"role,omitempty"`
}
