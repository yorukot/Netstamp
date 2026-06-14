package alert

import (
	"encoding/json"
	"net/http"

	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

func (h *Handler) handleListChannels(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	channels, err := h.service.ListChannels(
		r.Context(),
		appalert.ListChannelsInput{ProjectInput: projectInput(r, userID)},
	)
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"channels": channelResponses(channels)}, err)
}

func (h *Handler) handleCreateChannel(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body channelBody
	if decodeErr := httpx.DecodeJSON(r, &body); decodeErr != nil {
		httpx.WriteProblem(w, r, decodeErr)
		return
	}
	channel, err := h.service.CreateChannel(r.Context(), body.createInput(projectInput(r, userID)))
	writeJSONOrProblem(w, r, http.StatusCreated, map[string]any{"channel": channelResponse(channel)}, err)
}

func (h *Handler) handleGetChannel(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	channel, err := h.service.GetChannel(
		r.Context(),
		appalert.GetChannelInput{
			ProjectInput: projectInput(r, userID),
			ChannelID:    httpx.Path(r, "channel_id"),
		},
	)
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"channel": channelResponse(channel)}, err)
}

func (h *Handler) handleUpdateChannel(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body channelBody
	if decodeErr := httpx.DecodeJSON(r, &body); decodeErr != nil {
		httpx.WriteProblem(w, r, decodeErr)
		return
	}
	channel, err := h.service.UpdateChannel(r.Context(), body.updateInput(projectInput(r, userID), httpx.Path(r, "channel_id")))
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"channel": channelResponse(channel)}, err)
}

func (h *Handler) handleDeleteChannel(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	err = h.service.DeleteChannel(
		r.Context(),
		appalert.DeleteChannelInput{
			ProjectInput: projectInput(r, userID),
			ChannelID:    httpx.Path(r, "channel_id"),
		},
	)
	if err != nil {
		httpx.WriteProblem(w, r, mapAlertError(err, "delete notification channel failed"))
		return
	}
	httpx.WriteNoContent(w)
}

type channelBody struct {
	Name    string                  `json:"name"`
	Type    domainalert.ChannelType `json:"type"`
	Enabled bool                    `json:"enabled"`
	Config  json.RawMessage         `json:"config"`
}

func (b channelBody) createInput(project appalert.ProjectInput) appalert.CreateChannelInput {
	return appalert.CreateChannelInput{
		ProjectInput: project,
		Name:         b.Name,
		Type:         b.Type,
		Enabled:      b.Enabled,
		Config:       b.Config,
	}
}

func (b channelBody) updateInput(project appalert.ProjectInput, channelID string) appalert.UpdateChannelInput {
	return appalert.UpdateChannelInput{
		ProjectInput: project,
		ChannelID:    channelID,
		Name:         b.Name,
		Type:         b.Type,
		Enabled:      b.Enabled,
		Config:       b.Config,
	}
}
