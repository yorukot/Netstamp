package result

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service  *appresult.Service
	verifier appauth.TokenVerifier
}

func NewHandler(service *appresult.Service, verifier appauth.TokenVerifier) *Handler {
	return &Handler{
		service:  service,
		verifier: verifier,
	}
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireAuth(h.verifier))

		r.Get("/projects/{ref}/results/ping/series", h.handleQueryPingSeries)
		r.Get("/projects/{ref}/results/ping/insight", h.handleQueryPingInsight)
		r.Get("/projects/{ref}/results/latest", h.handleQueryLatestResults)
		r.Get("/projects/{ref}/results/tcp/series", h.handleQueryTCPSeries)
		r.Get("/projects/{ref}/results/tcp/insight", h.handleQueryTCPInsight)
		r.Get("/projects/{ref}/results/traceroute/runs", h.handleQueryTracerouteRuns)
		r.Get("/projects/{ref}/results/traceroute/insight", h.handleQueryTracerouteInsight)
		r.Get("/projects/{ref}/results/traceroute/topology", h.handleQueryTracerouteTopology)
	})
}

func (h *Handler) handleQueryPingSeries(w http.ResponseWriter, r *http.Request) {
	input, err := newQueryPingSeriesInput(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.queryPingSeries(r.Context(), input)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleQueryPingInsight(w http.ResponseWriter, r *http.Request) {
	input, err := newQueryPingInsightInput(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.queryPingInsight(r.Context(), input)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleQueryTCPSeries(w http.ResponseWriter, r *http.Request) {
	input, err := newQueryTCPSeriesInput(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.queryTCPSeries(r.Context(), input)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleQueryLatestResults(w http.ResponseWriter, r *http.Request) {
	input, err := newQueryLatestResultsInput(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.queryLatestResults(r.Context(), input)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleQueryTCPInsight(w http.ResponseWriter, r *http.Request) {
	input, err := newQueryTCPInsightInput(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.queryTCPInsight(r.Context(), input)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleQueryTracerouteRuns(w http.ResponseWriter, r *http.Request) {
	input, err := newQueryTracerouteRunsInput(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.queryTracerouteRuns(r.Context(), input)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleQueryTracerouteInsight(w http.ResponseWriter, r *http.Request) {
	input, err := newQueryTracerouteInsightInput(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.queryTracerouteInsight(r.Context(), input)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleQueryTracerouteTopology(w http.ResponseWriter, r *http.Request) {
	input, err := newQueryTracerouteTopologyInput(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.queryTracerouteTopology(r.Context(), input)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func newQueryPingSeriesInput(r *http.Request) (*queryPingSeriesInput, error) {
	from, err := httpx.QueryInt64(r, "from")
	if err != nil {
		return nil, err
	}
	to, err := httpx.QueryInt64(r, "to")
	if err != nil {
		return nil, err
	}
	maxDataPoints, err := httpx.QueryInt32(r, "maxDataPoints")
	if err != nil {
		return nil, err
	}
	return &queryPingSeriesInput{
		Ref:           httpx.Path(r, "ref"),
		ProbeID:       httpx.QueryString(r, "probeId"),
		CheckID:       httpx.QueryString(r, "checkId"),
		From:          from,
		To:            to,
		Series:        httpx.QueryString(r, "series"),
		MaxDataPoints: maxDataPoints,
	}, nil
}

func newQueryPingInsightInput(r *http.Request) (*queryPingInsightInput, error) {
	return newQueryInsightInput(r)
}

func newQueryTCPSeriesInput(r *http.Request) (*queryTCPSeriesInput, error) {
	from, err := httpx.QueryInt64(r, "from")
	if err != nil {
		return nil, err
	}
	to, err := httpx.QueryInt64(r, "to")
	if err != nil {
		return nil, err
	}
	maxDataPoints, err := httpx.QueryInt32(r, "maxDataPoints")
	if err != nil {
		return nil, err
	}
	return &queryTCPSeriesInput{
		Ref:           httpx.Path(r, "ref"),
		ProbeID:       httpx.QueryString(r, "probeId"),
		CheckID:       httpx.QueryString(r, "checkId"),
		From:          from,
		To:            to,
		Series:        httpx.QueryString(r, "series"),
		MaxDataPoints: maxDataPoints,
	}, nil
}

func newQueryLatestResultsInput(r *http.Request) (*queryLatestResultsInput, error) {
	return &queryLatestResultsInput{
		Ref:     httpx.Path(r, "ref"),
		ProbeID: httpx.QueryString(r, "probeId"),
		CheckID: httpx.QueryString(r, "checkId"),
		Type:    httpx.QueryString(r, "type"),
	}, nil
}

func newQueryTCPInsightInput(r *http.Request) (*queryTCPInsightInput, error) {
	return newQueryInsightInput(r)
}

func newQueryInsightInput(r *http.Request) (*queryInsightInput, error) {
	from, err := httpx.QueryInt64(r, "from")
	if err != nil {
		return nil, err
	}
	to, err := httpx.QueryInt64(r, "to")
	if err != nil {
		return nil, err
	}
	maxDataPoints, err := httpx.QueryInt32(r, "maxDataPoints")
	if err != nil {
		return nil, err
	}
	return &queryInsightInput{
		Ref:           httpx.Path(r, "ref"),
		ProbeID:       httpx.QueryString(r, "probeId"),
		CheckID:       httpx.QueryString(r, "checkId"),
		From:          from,
		To:            to,
		MaxDataPoints: maxDataPoints,
	}, nil
}

func newQueryTracerouteRunsInput(r *http.Request) (*queryTracerouteRunsInput, error) {
	from, err := httpx.QueryInt64(r, "from")
	if err != nil {
		return nil, err
	}
	to, err := httpx.QueryInt64(r, "to")
	if err != nil {
		return nil, err
	}
	limit, err := httpx.QueryInt32(r, "limit")
	if err != nil {
		return nil, err
	}
	cursor, err := httpx.QueryInt64(r, "cursor")
	if err != nil {
		return nil, err
	}
	return &queryTracerouteRunsInput{
		Ref:     httpx.Path(r, "ref"),
		ProbeID: httpx.QueryString(r, "probeId"),
		CheckID: httpx.QueryString(r, "checkId"),
		From:    from,
		To:      to,
		Limit:   limit,
		Cursor:  cursor,
	}, nil
}

func newQueryTracerouteInsightInput(r *http.Request) (*queryTracerouteInsightInput, error) {
	from, err := httpx.QueryInt64(r, "from")
	if err != nil {
		return nil, err
	}
	to, err := httpx.QueryInt64(r, "to")
	if err != nil {
		return nil, err
	}
	maxDataPoints, err := httpx.QueryInt32(r, "maxDataPoints")
	if err != nil {
		return nil, err
	}
	return &queryTracerouteInsightInput{
		Ref:           httpx.Path(r, "ref"),
		ProbeID:       httpx.QueryString(r, "probeId"),
		CheckID:       httpx.QueryString(r, "checkId"),
		From:          from,
		To:            to,
		MaxDataPoints: maxDataPoints,
	}, nil
}

func newQueryTracerouteTopologyInput(r *http.Request) (*queryTracerouteTopologyInput, error) {
	from, err := httpx.QueryInt64(r, "from")
	if err != nil {
		return nil, err
	}
	to, err := httpx.QueryInt64(r, "to")
	if err != nil {
		return nil, err
	}
	limit, err := httpx.QueryInt32(r, "limit")
	if err != nil {
		return nil, err
	}
	return &queryTracerouteTopologyInput{
		Ref:     httpx.Path(r, "ref"),
		ProbeID: httpx.QueryString(r, "probeId"),
		CheckID: httpx.QueryString(r, "checkId"),
		From:    from,
		To:      to,
		Limit:   limit,
	}, nil
}
