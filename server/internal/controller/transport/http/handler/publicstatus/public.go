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
