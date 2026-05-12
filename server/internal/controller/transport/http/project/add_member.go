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
		UserID:        input.Body.UserID,
		Role:          domainproject.Role(input.Body.Role),
	})
	if err != nil {
		return nil, mapProjectError(err, "add project member failed")
	}

	return &memberOutput{Body: memberOutputBody{Member: member}}, nil
}

type addMemberInput struct {
	Ref  string `path:"ref" minLength:"1" maxLength:"64" pattern:"^[a-z0-9-]+$" patternDescription:"lowercase letters, numbers, and dashes" doc:"Project UUID or slug." example:"engineering"`
	Body addMemberInputBody
}

type addMemberInputBody struct {
	UserID string `json:"userId,omitempty" doc:"User ID to add to the project."`
	Role   string `json:"role,omitempty" doc:"Project member role." example:"viewer"`
}
