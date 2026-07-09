package project

import (
	"context"
	"net/http"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func (h *Handler) handleCancelInvite(w http.ResponseWriter, r *http.Request) {
	_, err := h.cancelInvite(r.Context(), &cancelInviteInput{
		Ref:      httpx.Path(r, "ref"),
		InviteID: httpx.Path(r, "invite_id"),
	})
	writeNoContent(w, r, err)
}

func (h *Handler) cancelInvite(ctx context.Context, input *cancelInviteInput) (*cancelInviteOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	_, err = h.service.CancelInvite(ctx, appproject.CancelInviteInput{
		CurrentUserID: currentUserID,
		ProjectRef:    input.Ref,
		InviteID:      input.InviteID,
	})
	if err != nil {
		return nil, mapProjectError(err, "cancel project invite failed")
	}

	return &cancelInviteOutput{}, nil
}

type cancelInviteInput struct {
	Ref      string
	InviteID string
}

type cancelInviteOutput struct{}
