package publicstatus

import (
	"net/http"

	apppublic "github.com/yorukot/netstamp/internal/controller/application/publicstatus"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func (h *Handler) handleListPages(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r.Context())
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	pages, err := h.service.ListPages(r.Context(), apppublic.ListPagesInput{CurrentUserID: userID, ProjectRef: httpx.Path(r, "ref")})
	if err != nil {
		httpx.WriteProblem(w, r, mapPublicStatusError(err, "list public status pages failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, newPageListResponse(pages))
}

func (h *Handler) handleCreatePage(w http.ResponseWriter, r *http.Request) {
	var body pageInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	userID, err := currentUserID(r.Context())
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	page, err := h.service.CreatePage(r.Context(), apppublic.CreatePageInput{
		CurrentUserID:     userID,
		ProjectRef:        httpx.Path(r, "ref"),
		Slug:              body.Slug,
		Title:             body.Title,
		Description:       body.Description,
		Enabled:           defaultBool(body.Enabled, true),
		DefaultChartMode:  defaultPageChartMode(body.DefaultChartMode),
		DefaultChartRange: defaultChartRange(body.DefaultChartRange),
	})
	if err != nil {
		httpx.WriteProblem(w, r, mapPublicStatusError(err, "create public status page failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, pageResponseBody{Page: newPageBody(page, true)})
}

func (h *Handler) handleGetPage(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r.Context())
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	detail, err := h.service.GetPage(r.Context(), apppublic.GetPageInput{CurrentUserID: userID, ProjectRef: httpx.Path(r, "ref"), PageID: httpx.Path(r, "page_id")})
	if err != nil {
		httpx.WriteProblem(w, r, mapPublicStatusError(err, "get public status page failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, newPageDetailResponse(detail))
}

func (h *Handler) handleUpdatePage(w http.ResponseWriter, r *http.Request) {
	var body pageInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	userID, err := currentUserID(r.Context())
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	page, err := h.service.UpdatePage(r.Context(), apppublic.UpdatePageInput{
		CurrentUserID:     userID,
		ProjectRef:        httpx.Path(r, "ref"),
		PageID:            httpx.Path(r, "page_id"),
		Slug:              body.Slug,
		Title:             body.Title,
		Description:       body.Description,
		Enabled:           defaultBool(body.Enabled, true),
		DefaultChartMode:  defaultPageChartMode(body.DefaultChartMode),
		DefaultChartRange: defaultChartRange(body.DefaultChartRange),
	})
	if err != nil {
		httpx.WriteProblem(w, r, mapPublicStatusError(err, "update public status page failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, pageResponseBody{Page: newPageBody(page, true)})
}

func (h *Handler) handleDeletePage(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r.Context())
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	err = h.service.DeletePage(r.Context(), apppublic.DeletePageInput{CurrentUserID: userID, ProjectRef: httpx.Path(r, "ref"), PageID: httpx.Path(r, "page_id")})
	if err != nil {
		httpx.WriteProblem(w, r, mapPublicStatusError(err, "delete public status page failed"))
		return
	}
	httpx.WriteNoContent(w)
}
