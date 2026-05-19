package project

import (
	"context"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func (h *Handler) listMembers(ctx context.Context, input *projectRefInput) (*listMembersOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	members, err := h.service.ListMembers(ctx, appproject.ListMembersInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
	})
	if err != nil {
		return nil, mapProjectError(err, "list project members failed")
	}

	return &listMembersOutput{Body: listMembersOutputBody{Members: members}}, nil
}

type listMembersOutput struct {
	Body listMembersOutputBody
}

type listMembersOutputBody struct {
	Members []domainproject.Member `json:"members"`
}
