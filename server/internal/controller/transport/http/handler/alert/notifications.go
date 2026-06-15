package alert

import (
	"encoding/json"
	"net/http"

	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

func (h *Handler) handleListNotifications(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	notifications, err := h.service.ListNotifications(
		r.Context(),
		appalert.ListNotificationsInput{ProjectInput: projectInput(r, userID)},
	)
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"notifications": notificationResponses(notifications, h.emailSMTPConfigured)}, err)
}

func (h *Handler) handleCreateNotification(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body notificationBody
	if decodeErr := httpx.DecodeJSON(r, &body); decodeErr != nil {
		httpx.WriteProblem(w, r, decodeErr)
		return
	}
	notification, err := h.service.CreateNotification(r.Context(), body.createInput(projectInput(r, userID)))
	writeJSONOrProblem(w, r, http.StatusCreated, map[string]any{"notification": notificationResponse(notification, h.emailSMTPConfigured)}, err)
}

func (h *Handler) handleGetNotification(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	notification, err := h.service.GetNotification(
		r.Context(),
		appalert.GetNotificationInput{
			ProjectInput:   projectInput(r, userID),
			NotificationID: httpx.Path(r, "notification_id"),
		},
	)
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"notification": notificationResponse(notification, h.emailSMTPConfigured)}, err)
}

func (h *Handler) handleUpdateNotification(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body notificationBody
	if decodeErr := httpx.DecodeJSON(r, &body); decodeErr != nil {
		httpx.WriteProblem(w, r, decodeErr)
		return
	}
	notification, err := h.service.UpdateNotification(r.Context(), body.updateInput(projectInput(r, userID), httpx.Path(r, "notification_id")))
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"notification": notificationResponse(notification, h.emailSMTPConfigured)}, err)
}

func (h *Handler) handleDeleteNotification(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	err = h.service.DeleteNotification(
		r.Context(),
		appalert.DeleteNotificationInput{
			ProjectInput:   projectInput(r, userID),
			NotificationID: httpx.Path(r, "notification_id"),
		},
	)
	if err != nil {
		httpx.WriteProblem(w, r, mapAlertError(err, "delete notification failed"))
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) handleTestNotification(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	result, err := h.service.TestNotification(
		r.Context(),
		appalert.TestNotificationInput{
			ProjectInput:   projectInput(r, userID),
			NotificationID: httpx.Path(r, "notification_id"),
		},
	)
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"result": notificationTestResponse(result)}, err)
}

type notificationBody struct {
	Name    string                       `json:"name"`
	Type    domainalert.NotificationType `json:"type"`
	Enabled bool                         `json:"enabled"`
	Config  json.RawMessage              `json:"config"`
}

func (b notificationBody) createInput(project appalert.ProjectInput) appalert.CreateNotificationInput {
	return appalert.CreateNotificationInput{
		ProjectInput: project,
		Name:         b.Name,
		Type:         b.Type,
		Enabled:      b.Enabled,
		Config:       b.Config,
	}
}

func (b notificationBody) updateInput(project appalert.ProjectInput, notificationID string) appalert.UpdateNotificationInput {
	return appalert.UpdateNotificationInput{
		ProjectInput:   project,
		NotificationID: notificationID,
		Name:           b.Name,
		Type:           b.Type,
		Enabled:        b.Enabled,
		Config:         b.Config,
	}
}
