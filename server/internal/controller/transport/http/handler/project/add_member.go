package project

import (
	"context"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func (h *Handler) addMember(ctx context.Context, input *addMemberInput) (*memberOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	member, err := h.service.AddMember(ctx, appproject.AddMemberInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		Email:         input.Body.Email,
		Role:          domainproject.Role(input.Body.Role),
	})
	if err != nil {
		return nil, mapProjectError(err, "add project member failed")
	}

	return &memberOutput{Body: memberOutputBody{Member: member}}, nil
}

type addMemberInput struct {
	Ref  string
	Body addMemberInputBody
}

type addMemberInputBody struct {
	Email string `json:"email,omitempty"`
	Role  string `json:"role,omitempty"`
}
