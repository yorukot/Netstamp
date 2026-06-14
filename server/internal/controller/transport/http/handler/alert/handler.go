package alert

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Handler struct {
	service  *appalert.Service
	verifier appauth.TokenVerifier
}

func NewHandler(service *appalert.Service, verifier appauth.TokenVerifier) *Handler {
	return &Handler{service: service, verifier: verifier}
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireAuth(h.verifier))

		r.Get("/projects/{ref}/alerts/rules", h.handleListRules)
		r.Post("/projects/{ref}/alerts/rules", h.handleCreateRule)
		r.Get("/projects/{ref}/alerts/rules/{rule_id}", h.handleGetRule)
		r.Patch("/projects/{ref}/alerts/rules/{rule_id}", h.handleUpdateRule)
		r.Delete("/projects/{ref}/alerts/rules/{rule_id}", h.handleDeleteRule)

		r.Get("/projects/{ref}/alerts/incidents", h.handleListIncidents)
		r.Get("/projects/{ref}/alerts/incidents/{incident_id}", h.handleGetIncident)

		r.Get("/projects/{ref}/alerts/channels", h.handleListChannels)
		r.Post("/projects/{ref}/alerts/channels", h.handleCreateChannel)
		r.Get("/projects/{ref}/alerts/channels/{channel_id}", h.handleGetChannel)
		r.Patch("/projects/{ref}/alerts/channels/{channel_id}", h.handleUpdateChannel)
		r.Delete("/projects/{ref}/alerts/channels/{channel_id}", h.handleDeleteChannel)
	})
}

func (h *Handler) handleListRules(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var status *domainalert.RuleStatus
	if value := httpx.QueryString(r, "status"); value != "" {
		parsed := domainalert.RuleStatus(value)
		status = &parsed
	}
	var checkType *domaincheck.Type
	if value := httpx.QueryString(r, "checkType"); value != "" {
		parsed := domaincheck.Type(value)
		checkType = &parsed
	}
	rules, err := h.service.ListRules(r.Context(), appalert.ListRulesInput{ProjectInput: projectInput(r, userID), Status: status, CheckType: checkType})
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"rules": ruleResponses(rules)}, err)
}

func (h *Handler) handleCreateRule(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body ruleBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	rule, err := h.service.CreateRule(r.Context(), body.createInput(projectInput(r, userID)))
	writeJSONOrProblem(w, r, http.StatusCreated, map[string]any{"rule": ruleResponse(rule)}, err)
}

func (h *Handler) handleGetRule(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	rule, err := h.service.GetRule(r.Context(), appalert.GetRuleInput{ProjectInput: projectInput(r, userID), RuleID: httpx.Path(r, "rule_id")})
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"rule": ruleResponse(rule)}, err)
}

func (h *Handler) handleUpdateRule(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body ruleBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	rule, err := h.service.UpdateRule(r.Context(), body.updateInput(projectInput(r, userID), httpx.Path(r, "rule_id")))
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"rule": ruleResponse(rule)}, err)
}

func (h *Handler) handleDeleteRule(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	err = h.service.DeleteRule(r.Context(), appalert.DeleteRuleInput{ProjectInput: projectInput(r, userID), RuleID: httpx.Path(r, "rule_id")})
	if err != nil {
		httpx.WriteProblem(w, r, mapAlertError(err, "delete alert rule failed"))
		return
	}
	httpx.WriteNoContent(w)
}

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
	incidents, err := h.service.ListIncidents(r.Context(), appalert.ListIncidentsInput{ProjectInput: projectInput(r, userID), Status: status, Limit: limit})
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"incidents": incidentResponses(incidents)}, err)
}

func (h *Handler) handleGetIncident(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	incident, err := h.service.GetIncident(r.Context(), appalert.GetIncidentInput{ProjectInput: projectInput(r, userID), IncidentID: httpx.Path(r, "incident_id")})
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"incident": incidentResponse(incident)}, err)
}

func (h *Handler) handleListChannels(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	channels, err := h.service.ListChannels(r.Context(), appalert.ListChannelsInput{ProjectInput: projectInput(r, userID)})
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"channels": channelResponses(channels)}, err)
}

func (h *Handler) handleCreateChannel(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body channelBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	channel, err := h.service.CreateChannel(r.Context(), body.createInput(projectInput(r, userID)))
	writeJSONOrProblem(w, r, http.StatusCreated, map[string]any{"channel": channelResponse(channel)}, err)
}

func (h *Handler) handleGetChannel(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	channel, err := h.service.GetChannel(r.Context(), appalert.GetChannelInput{ProjectInput: projectInput(r, userID), ChannelID: httpx.Path(r, "channel_id")})
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"channel": channelResponse(channel)}, err)
}

func (h *Handler) handleUpdateChannel(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body channelBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	channel, err := h.service.UpdateChannel(r.Context(), body.updateInput(projectInput(r, userID), httpx.Path(r, "channel_id")))
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"channel": channelResponse(channel)}, err)
}

func (h *Handler) handleDeleteChannel(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	err = h.service.DeleteChannel(r.Context(), appalert.DeleteChannelInput{ProjectInput: projectInput(r, userID), ChannelID: httpx.Path(r, "channel_id")})
	if err != nil {
		httpx.WriteProblem(w, r, mapAlertError(err, "delete notification channel failed"))
		return
	}
	httpx.WriteNoContent(w)
}

type ruleBody struct {
	Name                   string                   `json:"name"`
	Description            *string                  `json:"description"`
	Enabled                bool                     `json:"enabled"`
	Severity               domainalert.Severity     `json:"severity"`
	Scope                  ruleScopeBody            `json:"scope"`
	Condition              alertcondition.Condition `json:"condition"`
	CooldownSeconds        int32                    `json:"cooldownSeconds"`
	NotificationChannelIDs []string                 `json:"notificationChannelIds"`
}

type ruleScopeBody struct {
	CheckType domaincheck.Type `json:"checkType"`
	ProbeID   *string          `json:"probeId"`
	CheckID   *string          `json:"checkId"`
}

func (b ruleBody) createInput(project appalert.ProjectInput) appalert.CreateRuleInput {
	return appalert.CreateRuleInput{
		ProjectInput: project, Name: b.Name, Description: b.Description, Enabled: b.Enabled, Severity: b.Severity,
		CheckType: b.Scope.CheckType, ProbeID: b.Scope.ProbeID, CheckID: b.Scope.CheckID, Condition: b.Condition,
		CooldownSeconds: b.CooldownSeconds, NotificationChannelIDs: b.NotificationChannelIDs,
	}
}

func (b ruleBody) updateInput(project appalert.ProjectInput, ruleID string) appalert.UpdateRuleInput {
	return appalert.UpdateRuleInput{
		ProjectInput: project, RuleID: ruleID, Name: b.Name, Description: b.Description, Enabled: b.Enabled, Severity: b.Severity,
		CheckType: b.Scope.CheckType, ProbeID: b.Scope.ProbeID, CheckID: b.Scope.CheckID, Condition: b.Condition,
		CooldownSeconds: b.CooldownSeconds, NotificationChannelIDs: b.NotificationChannelIDs,
	}
}

type channelBody struct {
	Name    string                  `json:"name"`
	Type    domainalert.ChannelType `json:"type"`
	Enabled bool                    `json:"enabled"`
	Config  json.RawMessage         `json:"config"`
}

func (b channelBody) createInput(project appalert.ProjectInput) appalert.CreateChannelInput {
	return appalert.CreateChannelInput{ProjectInput: project, Name: b.Name, Type: b.Type, Enabled: b.Enabled, Config: b.Config}
}

func (b channelBody) updateInput(project appalert.ProjectInput, channelID string) appalert.UpdateChannelInput {
	return appalert.UpdateChannelInput{ProjectInput: project, ChannelID: channelID, Name: b.Name, Type: b.Type, Enabled: b.Enabled, Config: b.Config}
}

func currentUserID(r *http.Request) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(r.Context())
	if !ok || claims.Subject == "" {
		return "", httpx.Unauthorized("missing auth cookie")
	}
	return claims.Subject, nil
}

func projectInput(r *http.Request, userID string) appalert.ProjectInput {
	return appalert.ProjectInput{ProjectRef: httpx.Path(r, "ref"), CurrentUserID: userID}
}

func writeJSONOrProblem(w http.ResponseWriter, r *http.Request, status int, body any, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, mapAlertError(err, "alert request failed"))
		return
	}
	httpx.WriteJSON(w, status, body)
}

func mapAlertError(err error, fallback string) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, domainproject.ErrMemberNotFound), errors.Is(err, identity.ErrUserNotFound), errors.Is(err, domainalert.ErrRuleNotFound), errors.Is(err, domainalert.ErrIncidentNotFound), errors.Is(err, domainalert.ErrChannelNotFound):
		return httpx.NotFound("not found")
	case errors.Is(err, appalert.ErrForbidden):
		return httpx.Forbidden("forbidden")
	case errors.Is(err, appalert.ErrInvalidInput), errors.Is(err, domainalert.ErrInvalidInput), errors.Is(err, alertcondition.ErrInvalidCondition), errors.Is(err, domaincheck.ErrInvalidInput):
		return httpx.UnprocessableEntity("invalid alert input")
	default:
		return httpx.InternalServerError(fallback)
	}
}
