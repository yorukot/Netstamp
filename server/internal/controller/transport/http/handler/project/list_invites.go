package project

import (
	"context"
	"net/http"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func (h *Handler) handleListProjectInvites(w http.ResponseWriter, r *http.Request) {
	output, err := h.listProjectInvites(r.Context(), &projectRefInput{Ref: httpx.Path(r, "ref")})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleListUserInvites(w http.ResponseWriter, r *http.Request) {
	output, err := h.listUserInvites(r.Context())
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) listProjectInvites(ctx context.Context, input *projectRefInput) (*inviteListOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	invites, err := h.service.ListProjectInvites(ctx, appproject.ListProjectInvitesInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
	})
	if err != nil {
		return nil, mapProjectError(err, "list project invites failed")
	}

	return &inviteListOutput{Body: inviteListOutputBody{Invites: invites}}, nil
}

func (h *Handler) listUserInvites(ctx context.Context) (*inviteListOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	invites, err := h.service.ListUserInvites(ctx, appproject.ListUserInvitesInput{
		CurrentUserID: currentUserID,
	})
	if err != nil {
		return nil, mapProjectError(err, "list user invites failed")
	}

	return &inviteListOutput{Body: inviteListOutputBody{Invites: invites}}, nil
}
