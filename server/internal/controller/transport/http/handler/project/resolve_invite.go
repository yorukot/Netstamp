package project

import (
	"context"
	"net/http"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func (h *Handler) handleAcceptInvite(w http.ResponseWriter, r *http.Request) {
	output, err := h.acceptInvite(r.Context(), &resolveInviteInput{InviteID: httpx.Path(r, "invite_id")})
	writeInviteOutput(w, r, http.StatusOK, output, err)
}

func (h *Handler) handleRejectInvite(w http.ResponseWriter, r *http.Request) {
	output, err := h.rejectInvite(r.Context(), &resolveInviteInput{InviteID: httpx.Path(r, "invite_id")})
	writeInviteOutput(w, r, http.StatusOK, output, err)
}

func (h *Handler) acceptInvite(ctx context.Context, input *resolveInviteInput) (*inviteOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	invite, err := h.service.AcceptInvite(ctx, appproject.ResolveInviteInput{
		CurrentUserID: currentUserID,
		InviteID:      input.InviteID,
	})
	if err != nil {
		return nil, mapProjectError(err, "accept project invite failed")
	}

	return &inviteOutput{Body: inviteOutputBody{Invite: invite}}, nil
}

func (h *Handler) rejectInvite(ctx context.Context, input *resolveInviteInput) (*inviteOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	invite, err := h.service.RejectInvite(ctx, appproject.ResolveInviteInput{
		CurrentUserID: currentUserID,
		InviteID:      input.InviteID,
	})
	if err != nil {
		return nil, mapProjectError(err, "reject project invite failed")
	}

	return &inviteOutput{Body: inviteOutputBody{Invite: invite}}, nil
}

type resolveInviteInput struct {
	InviteID string
}
