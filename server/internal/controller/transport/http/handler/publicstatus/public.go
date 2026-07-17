package publicstatus

import (
	"net/http"

	apppublic "github.com/yorukot/netstamp/internal/controller/application/publicstatus"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

func (h *Handler) handleGetPublicStatusSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.service.GetPublicSummary(r.Context(), apppublic.PublicSummaryInput{
		Slug: httpx.Path(r, "slug"),
	})
	if err != nil {
		httpx.WriteProblem(w, r, mapPublicStatusError(err, "get public status summary failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, newPublicStatusSummaryResponse(summary))
}

func (h *Handler) handleGetPublicStatusEditorContext(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r.Context())
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	context, err := h.service.GetEditorContext(r.Context(), apppublic.EditorContextInput{
		CurrentUserID: userID,
		Slug:          httpx.Path(r, "slug"),
	})
	if err != nil {
		httpx.WriteProblem(w, r, mapPublicStatusError(err, "get public status editor context failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, publicStatusEditorContextResponseBody{ProjectRef: context.ProjectRef, PageID: context.PageID})
}

func (h *Handler) handleGetPublicStatusElements(w http.ResponseWriter, r *http.Request) {
	elements, err := h.service.GetPublicElements(r.Context(), apppublic.PublicElementsInput{
		Slug: httpx.Path(r, "slug"),
	})
	if err != nil {
		httpx.WriteProblem(w, r, mapPublicStatusError(err, "get public status elements failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, newPublicStatusElementsResponse(elements))
}

func (h *Handler) handleGetPublicStatusIncidents(w http.ResponseWriter, r *http.Request) {
	limit, err := httpx.QueryInt32(r, "limit")
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	incidents, err := h.service.GetPublicIncidents(r.Context(), apppublic.PublicIncidentsInput{
		Slug:  httpx.Path(r, "slug"),
		Limit: limit,
	})
	if err != nil {
		httpx.WriteProblem(w, r, mapPublicStatusError(err, "get public status incidents failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, newPublicStatusIncidentsResponse(incidents))
}

func (h *Handler) handleGetPublicStatusElementChart(w http.ResponseWriter, r *http.Request) {
	chartRange, err := queryChartRange(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	chart, err := h.service.GetPublicElementChart(r.Context(), apppublic.PublicElementChartInput{
		Slug:      httpx.Path(r, "slug"),
		ElementID: httpx.Path(r, "element_id"),
		Range:     chartRange,
	})
	if err != nil {
		httpx.WriteProblem(w, r, mapPublicStatusError(err, "get public status element chart failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, newPublicStatusElementChartResponse(chart))
}

func (h *Handler) handleGetPublicStatusElementDailyStatus(w http.ResponseWriter, r *http.Request) {
	dailyStatusRange, err := queryDailyStatusRange(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	dailyStatus, err := h.service.GetPublicElementDailyStatus(r.Context(), apppublic.PublicElementDailyStatusInput{
		Slug:      httpx.Path(r, "slug"),
		ElementID: httpx.Path(r, "element_id"),
		Range:     dailyStatusRange,
	})
	if err != nil {
		httpx.WriteProblem(w, r, mapPublicStatusError(err, "get public status element daily status failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, newPublicStatusElementDailyStatusResponse(dailyStatus))
}

func queryChartRange(r *http.Request) (*domainpublic.ChartRange, error) {
	value := httpx.QueryString(r, "range")
	if value == "" {
		return nil, nil //nolint:nilnil // Nil means use configured chart range.
	}
	chartRange := domainpublic.ChartRange(value)
	normalized, err := domainpublic.VNChartRange(chartRange)
	if err != nil {
		return nil, httpx.BadRequest("invalid query parameter range")
	}
	return &normalized, nil
}

func queryDailyStatusRange(r *http.Request) (*domainpublic.ChartRange, error) {
	value := httpx.QueryString(r, "range")
	if value == "" {
		return nil, nil //nolint:nilnil // Nil means use the only supported daily status range.
	}
	if value != string(domainpublic.ChartRange30d) {
		return nil, httpx.BadRequest("invalid query parameter range")
	}
	dailyStatusRange := domainpublic.ChartRange30d
	return &dailyStatusRange, nil
}
