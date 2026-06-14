package alert

import (
	"net/http"

	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
)

func (h *Handler) handleListIncidents(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var status *domainalert.IncidentStatus
	if value := httpx.QueryString(r, "status"); value != "" {
		parsed := domainalert.IncidentStatus(value)
		status = &parsed
	}
	limit, err := httpx.QueryInt32(r, "limit")
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	incidents, err := h.service.ListIncidents(
		r.Context(),
		appalert.ListIncidentsInput{
			ProjectInput: projectInput(r, userID),
			Status:       status,
			Limit:        limit,
		},
	)
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"incidents": incidentResponses(incidents)}, err)
}

func (h *Handler) handleGetIncident(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	incident, err := h.service.GetIncident(
		r.Context(),
		appalert.GetIncidentInput{
			ProjectInput: projectInput(r, userID),
			IncidentID:   httpx.Path(r, "incident_id"),
		},
	)
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"incident": incidentResponse(incident)}, err)
}
