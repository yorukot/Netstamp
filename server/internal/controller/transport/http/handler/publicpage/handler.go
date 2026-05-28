package publicpage

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	apppublicpage "github.com/yorukot/netstamp/internal/controller/application/publicpage"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service  *apppublicpage.Service
	verifier appauth.TokenVerifier
}

func NewHandler(service *apppublicpage.Service, verifier appauth.TokenVerifier) *Handler {
	return &Handler{
		service:  service,
		verifier: verifier,
	}
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireAuth(h.verifier))

		r.Get("/projects/{ref}/public-pages", h.handleListPages)
		r.Post("/projects/{ref}/public-pages", h.handleCreatePage)
		r.Get("/projects/{ref}/public-pages/{page_id}", h.handleGetPage)
		r.Patch("/projects/{ref}/public-pages/{page_id}", h.handleUpdatePage)
		r.Delete("/projects/{ref}/public-pages/{page_id}", h.handleDeletePage)
		r.Post("/projects/{ref}/public-pages/{page_id}/folders", h.handleCreateFolder)
		r.Patch("/projects/{ref}/public-pages/{page_id}/folders/{folder_id}", h.handleUpdateFolder)
		r.Delete("/projects/{ref}/public-pages/{page_id}/folders/{folder_id}", h.handleDeleteFolder)
		r.Put("/projects/{ref}/public-pages/{page_id}/folders/{folder_id}/checks", h.handleSetFolderChecks)
	})

	api.Get("/public-pages/{slug}", h.handleGetPublicPage)
	api.Get("/public-pages/{slug}/results/ping/insight", h.handleQueryPublicPingInsight)
}

func (h *Handler) handleListPages(w http.ResponseWriter, r *http.Request) {
	output, err := h.listPages(r)
	writePageListOutput(w, r, output, err)
}

func (h *Handler) handleCreatePage(w http.ResponseWriter, r *http.Request) {
	var body createPageInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.createPage(r, body)
	writePageOutput(w, r, http.StatusCreated, output, err)
}

func (h *Handler) handleGetPage(w http.ResponseWriter, r *http.Request) {
	output, err := h.getPage(r)
	writePageOutput(w, r, http.StatusOK, output, err)
}

func (h *Handler) handleUpdatePage(w http.ResponseWriter, r *http.Request) {
	var body updatePageInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.updatePage(r, body)
	writePageOutput(w, r, http.StatusOK, output, err)
}

func (h *Handler) handleDeletePage(w http.ResponseWriter, r *http.Request) {
	err := h.deletePage(r)
	writeNoContent(w, r, err)
}

func (h *Handler) handleCreateFolder(w http.ResponseWriter, r *http.Request) {
	var body createFolderInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.createFolder(r, body)
	writeFolderOutput(w, r, http.StatusCreated, output, err)
}

func (h *Handler) handleUpdateFolder(w http.ResponseWriter, r *http.Request) {
	var body updateFolderInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.updateFolder(r, body)
	writeFolderOutput(w, r, http.StatusOK, output, err)
}

func (h *Handler) handleDeleteFolder(w http.ResponseWriter, r *http.Request) {
	err := h.deleteFolder(r)
	writeNoContent(w, r, err)
}

func (h *Handler) handleSetFolderChecks(w http.ResponseWriter, r *http.Request) {
	var body setFolderChecksInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.setFolderChecks(r, body)
	writeChecksOutput(w, r, output, err)
}

func (h *Handler) handleGetPublicPage(w http.ResponseWriter, r *http.Request) {
	output, err := h.getPublicPage(r)
	writePageOutput(w, r, http.StatusOK, output, err)
}

func (h *Handler) handleQueryPublicPingInsight(w http.ResponseWriter, r *http.Request) {
	output, err := h.queryPublicPingInsight(r)
	writePingInsightOutput(w, r, output, err)
}

func writePageOutput(w http.ResponseWriter, r *http.Request, status int, output *pageOutput, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, status, output.Body)
}

func writePageListOutput(w http.ResponseWriter, r *http.Request, output *pageListOutput, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func writeFolderOutput(w http.ResponseWriter, r *http.Request, status int, output *folderOutput, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, status, output.Body)
}

func writeChecksOutput(w http.ResponseWriter, r *http.Request, output *checksOutput, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func writePingInsightOutput(w http.ResponseWriter, r *http.Request, output *pingInsightOutput, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func writeNoContent(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteNoContent(w)
}
