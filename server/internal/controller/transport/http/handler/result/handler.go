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
	api.With(httpmiddleware.RequireAuth(h.verifier)).Get("/projects/{ref}/results/ping/series", h.handleQueryPingSeries)
	api.With(httpmiddleware.RequireAuth(h.verifier)).Get("/projects/{ref}/results/traceroute/runs", h.handleQueryTracerouteRuns)
	api.With(httpmiddleware.RequireAuth(h.verifier)).Get("/projects/{ref}/results/traceroute/topology", h.handleQueryTracerouteTopology)
	api.With(httpmiddleware.RequireAuth(h.verifier)).Get("/projects/{ref}/measurements", h.handleQueryMeasurements)
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

func (h *Handler) handleQueryMeasurements(w http.ResponseWriter, r *http.Request) {
	input, err := newQueryMeasurementsInput(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.queryMeasurements(r.Context(), input)
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
		Metric:        httpx.QueryString(r, "metric"),
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

func newQueryMeasurementsInput(r *http.Request) (*queryMeasurementsInput, error) {
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
	return &queryMeasurementsInput{
		Ref:     httpx.Path(r, "ref"),
		ProbeID: httpx.QueryString(r, "probeId"),
		CheckID: httpx.QueryString(r, "checkId"),
		Type:    httpx.QueryString(r, "type"),
		Status:  httpx.QueryString(r, "status"),
		From:    from,
		To:      to,
		Limit:   limit,
		Cursor:  cursor,
	}, nil
}
