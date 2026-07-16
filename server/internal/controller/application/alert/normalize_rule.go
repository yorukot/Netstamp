package alert

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

func normalizeCreateRule(projectID string, input CreateRuleInput) (domainalert.Rule, error) {
	status := domainalert.RuleStatusDisabled
	if input.Enabled {
		status = domainalert.RuleStatusEnabled
	}
	return normalizeRule(domainalert.Rule{ProjectID: projectID, CreatedByUserID: input.CurrentUserID, Status: status}, input.Name, input.Description, input.Severity, input.CheckType, input.ProbeID, input.CheckID, input.Condition, input.TriggerAfterSeconds, input.CooldownSeconds, input.NotificationIDs)
}

func normalizeUpdateRule(projectID string, input UpdateRuleInput) (domainalert.Rule, error) {
	status := domainalert.RuleStatusDisabled
	if input.Enabled {
		status = domainalert.RuleStatusEnabled
	}
	base := domainalert.Rule{ProjectID: projectID, ID: input.RuleID, CreatedByUserID: input.CurrentUserID, Status: status}
	return normalizeRule(base, input.Name, input.Description, input.Severity, input.CheckType, input.ProbeID, input.CheckID, input.Condition, input.TriggerAfterSeconds, input.CooldownSeconds, input.NotificationIDs)
}

func normalizeRule(base domainalert.Rule, name string, description *string, severity domainalert.Severity, checkType domaincheck.Type, probeID, checkID *string, condition alertcondition.Condition, triggerAfterSeconds, cooldownSeconds int32, notificationIDs []string) (domainalert.Rule, error) {
	var err error
	base.Name, err = domainalert.VNRuleName(name)
	if err != nil {
		return domainalert.Rule{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.Description, err = domainalert.VNDescription(description)
	if err != nil {
		return domainalert.Rule{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.Severity, err = domainalert.VNSeverity(severity)
	if err != nil {
		return domainalert.Rule{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.CheckType, err = domaincheck.VNCheckType(checkType)
	if err != nil || base.CheckType == domaincheck.TypeTraceroute {
		return domainalert.Rule{}, fmt.Errorf("%w: unsupported check type", ErrInvalidInput)
	}
	base.Condition, err = alertcondition.Validate(condition)
	if err != nil {
		return domainalert.Rule{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	if !alertcondition.CompatibleWithCheckType(base.Condition.Metric, base.CheckType) {
		return domainalert.Rule{}, fmt.Errorf("%w: condition metric is not compatible with check type", ErrInvalidInput)
	}
	base.ConditionJSON, err = alertcondition.CanonicalJSON(base.Condition)
	if err != nil {
		return domainalert.Rule{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.ConditionVersion = "metric_threshold.v1"
	base.TriggerAfterSeconds, err = domainalert.VNTriggerAfterSeconds(triggerAfterSeconds)
	if err != nil {
		return domainalert.Rule{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.CooldownSeconds, err = domainalert.VNCooldownSeconds(cooldownSeconds)
	if err != nil {
		return domainalert.Rule{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	base.ProbeID = probeID
	base.CheckID = checkID
	if err := validateUUIDPtr(probeID); err != nil {
		return domainalert.Rule{}, err
	}
	if err := validateUUIDPtr(checkID); err != nil {
		return domainalert.Rule{}, err
	}
	base.ProbeSelector = json.RawMessage(`{}`)
	for _, notificationID := range notificationIDs {
		if _, parseErr := uuid.Parse(notificationID); parseErr != nil {
			return domainalert.Rule{}, fmt.Errorf("%w: invalid notification id", ErrInvalidInput)
		}
	}
	base.NotificationIDs = append([]string{}, notificationIDs...)
	return base, nil
}
