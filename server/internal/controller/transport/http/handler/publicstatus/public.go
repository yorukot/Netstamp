package publicstatus

import (
	"net/http"
	"strconv"

	apppublic "github.com/yorukot/netstamp/internal/controller/application/publicstatus"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

func (h *Handler) handleGetPublicStatusPage(w http.ResponseWriter, r *http.Request) {
	includeCharts, err := queryBool(r, "includeCharts", true)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	chartRange, err := queryChartRange(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	rendered, err := h.service.GetPublicPage(r.Context(), apppublic.PublicPageInput{
		Slug:          httpx.Path(r, "slug"),
		IncludeCharts: includeCharts,
		Range:         chartRange,
	})
	if err != nil {
		httpx.WriteProblem(w, r, mapPublicStatusError(err, "get public status page failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, newPublicStatusPageResponse(rendered))
}

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

func queryBool(r *http.Request, name string, fallback bool) (bool, error) {
	value := httpx.QueryString(r, name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, httpx.BadRequest("invalid query parameter " + name)
	}
	return parsed, nil
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
