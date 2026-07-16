package alert

import (
	"net/http"

	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

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
	rules, err := h.service.ListRules(
		r.Context(),
		appalert.ListRulesInput{
			ProjectInput: projectInput(r, userID),
			Status:       status,
			CheckType:    checkType,
		},
	)
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"rules": ruleResponses(rules)}, err)
}

func (h *Handler) handleCreateRule(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body ruleBody
	if decodeErr := httpx.DecodeJSON(r, &body); decodeErr != nil {
		httpx.WriteProblem(w, r, decodeErr)
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
	rule, err := h.service.GetRule(
		r.Context(),
		appalert.GetRuleInput{
			ProjectInput: projectInput(r, userID),
			RuleID:       httpx.Path(r, "rule_id"),
		},
	)
	writeJSONOrProblem(w, r, http.StatusOK, map[string]any{"rule": ruleResponse(rule)}, err)
}

func (h *Handler) handleUpdateRule(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body ruleBody
	if decodeErr := httpx.DecodeJSON(r, &body); decodeErr != nil {
		httpx.WriteProblem(w, r, decodeErr)
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
	err = h.service.DeleteRule(
		r.Context(),
		appalert.DeleteRuleInput{
			ProjectInput: projectInput(r, userID),
			RuleID:       httpx.Path(r, "rule_id"),
		},
	)
	if err != nil {
		httpx.WriteProblem(w, r, mapAlertError(err, "delete alert rule failed"))
		return
	}
	httpx.WriteNoContent(w)
}

type ruleBody struct {
	Name                string                   `json:"name"`
	Description         *string                  `json:"description"`
	Enabled             bool                     `json:"enabled"`
	Severity            domainalert.Severity     `json:"severity"`
	Scope               ruleScopeBody            `json:"scope"`
	Condition           alertcondition.Condition `json:"condition"`
	TriggerAfterSeconds int32                    `json:"triggerAfterSeconds"`
	CooldownSeconds     int32                    `json:"cooldownSeconds"`
	NotificationIDs     []string                 `json:"notificationIds"`
}

type ruleScopeBody struct {
	CheckType domaincheck.Type `json:"checkType"`
	ProbeID   *string          `json:"probeId"`
	CheckID   *string          `json:"checkId"`
}

func (b ruleBody) createInput(project appalert.ProjectInput) appalert.CreateRuleInput {
	return appalert.CreateRuleInput{
		ProjectInput:        project,
		Name:                b.Name,
		Description:         b.Description,
		Enabled:             b.Enabled,
		Severity:            b.Severity,
		CheckType:           b.Scope.CheckType,
		ProbeID:             b.Scope.ProbeID,
		CheckID:             b.Scope.CheckID,
		Condition:           b.Condition,
		TriggerAfterSeconds: b.TriggerAfterSeconds,
		CooldownSeconds:     b.CooldownSeconds,
		NotificationIDs:     b.NotificationIDs,
	}
}

func (b ruleBody) updateInput(project appalert.ProjectInput, ruleID string) appalert.UpdateRuleInput {
	return appalert.UpdateRuleInput{
		ProjectInput:        project,
		RuleID:              ruleID,
		Name:                b.Name,
		Description:         b.Description,
		Enabled:             b.Enabled,
		Severity:            b.Severity,
		CheckType:           b.Scope.CheckType,
		ProbeID:             b.Scope.ProbeID,
		CheckID:             b.Scope.CheckID,
		Condition:           b.Condition,
		TriggerAfterSeconds: b.TriggerAfterSeconds,
		CooldownSeconds:     b.CooldownSeconds,
		NotificationIDs:     b.NotificationIDs,
	}
}
