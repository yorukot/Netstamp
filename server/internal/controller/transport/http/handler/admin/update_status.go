package admin

import (
	"net/http"
	"time"

	appadmin "github.com/yorukot/netstamp/internal/controller/application/admin"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

type updateStatusResponseBody struct {
	CurrentVersion  string     `json:"currentVersion"`
	LatestVersion   *string    `json:"latestVersion"`
	UpdateAvailable bool       `json:"updateAvailable"`
	ReleaseURL      *string    `json:"releaseUrl"`
	PublishedAt     *time.Time `json:"publishedAt"`
	LastCheckedAt   *time.Time `json:"lastCheckedAt"`
	CheckError      *string    `json:"checkError"`
}

func (h *Handler) handleGetUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("admin service is unavailable"))
		return
	}
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	status, err := h.service.GetUpdateStatus(r.Context(), appadmin.UpdateStatusInput{CurrentUserID: userID})
	if err != nil {
		httpx.WriteProblem(w, r, mapAdminError(err, "get update status failed"))
		return
	}
	w.Header().Set("Cache-Control", settingsCacheControl)
	httpx.WriteJSON(w, http.StatusOK, updateStatusResponseBody{
		CurrentVersion: status.CurrentVersion, LatestVersion: status.LatestVersion,
		UpdateAvailable: status.UpdateAvailable, ReleaseURL: status.ReleaseURL,
		PublishedAt: status.PublishedAt, LastCheckedAt: status.LastCheckedAt, CheckError: status.CheckError,
	})
}
