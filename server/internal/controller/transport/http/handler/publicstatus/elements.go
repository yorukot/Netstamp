package publicstatus

import (
	"net/http"

	apppublic "github.com/yorukot/netstamp/internal/controller/application/publicstatus"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func (h *Handler) handleCreateElement(w http.ResponseWriter, r *http.Request) {
	var body elementInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	userID, err := currentUserID(r.Context())
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	element, err := h.service.CreateElement(r.Context(), apppublic.CreateElementInput{
		CurrentUserID:           userID,
		ProjectRef:              httpx.Path(r, "ref"),
		PageID:                  httpx.Path(r, "page_id"),
		ParentElementID:         body.ParentElementID,
		Kind:                    body.Kind,
		CheckID:                 body.CheckID,
		AssignmentSelectionMode: body.AssignmentSelectionMode,
		AssignmentIDs:           body.AssignmentIDs,
		Title:                   body.Title,
		Description:             body.Description,
		SortOrder:               body.SortOrder,
		DisplayMode:             defaultElementDisplayMode(body.DisplayMode),
		ChartMode:               defaultElementChartMode(body.ChartMode),
		ChartRange:              body.ChartRange,
	})
	if err != nil {
		httpx.WriteProblem(w, r, mapPublicStatusError(err, "create public status page element failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, elementResponseBody{Element: newElementBody(element)})
}

func (h *Handler) handleUpdateElement(w http.ResponseWriter, r *http.Request) {
	var body elementInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	userID, err := currentUserID(r.Context())
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	element, err := h.service.UpdateElement(r.Context(), apppublic.UpdateElementInput{
		CurrentUserID:           userID,
		ProjectRef:              httpx.Path(r, "ref"),
		PageID:                  httpx.Path(r, "page_id"),
		ElementID:               httpx.Path(r, "element_id"),
		ParentElementID:         body.ParentElementID,
		Kind:                    body.Kind,
		CheckID:                 body.CheckID,
		AssignmentSelectionMode: body.AssignmentSelectionMode,
		AssignmentIDs:           body.AssignmentIDs,
		Title:                   body.Title,
		Description:             body.Description,
		SortOrder:               body.SortOrder,
		DisplayMode:             defaultElementDisplayMode(body.DisplayMode),
		ChartMode:               defaultElementChartMode(body.ChartMode),
		ChartRange:              body.ChartRange,
	})
	if err != nil {
		httpx.WriteProblem(w, r, mapPublicStatusError(err, "update public status page element failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, elementResponseBody{Element: newElementBody(element)})
}

func (h *Handler) handleDeleteElement(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r.Context())
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	err = h.service.DeleteElement(r.Context(), apppublic.DeleteElementInput{
		CurrentUserID: userID,
		ProjectRef:    httpx.Path(r, "ref"),
		PageID:        httpx.Path(r, "page_id"),
		ElementID:     httpx.Path(r, "element_id"),
	})
	if err != nil {
		httpx.WriteProblem(w, r, mapPublicStatusError(err, "delete public status page element failed"))
		return
	}
	httpx.WriteNoContent(w)
}
