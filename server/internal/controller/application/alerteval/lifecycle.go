package alerteval

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	"github.com/yorukot/netstamp/internal/domain/alertcondition"
)

// ReconcileStoppedEvaluations resolves incidents and clears pending state for
// alert targets that can no longer be evaluated under the current project state.
func (s *Service) ReconcileStoppedEvaluations(ctx context.Context, projectID string) error {
	ctx, flow := s.startLifecycleFlow(ctx, projectID)
	defer flow.end()

	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		candidates, err := s.repo.ListInactiveActiveIncidents(ctx, projectID)
		if err != nil {
			return flow.failure(AlertEvalEventLifecycleReconcileFailure, AlertEvalReasonInactiveIncidentListFail, err)
		}

		now := s.now()
		for _, candidate := range candidates {
			flow.ruleID = candidate.Rule.ID
			flow.probeID = candidate.Incident.ProbeID
			flow.checkID = candidate.Incident.CheckID
			flow.checkType = string(candidate.Incident.CheckType)
			flow.setIncidentID(candidate.Incident.ID)

			incident, resolveErr := s.repo.ResolveInactiveIncident(ctx, candidate.Incident.ID, now)
			if errors.Is(resolveErr, domainalert.ErrIncidentNotFound) {
				continue
			}
			if resolveErr != nil {
				return flow.failure(AlertEvalEventLifecycleReconcileFailure, AlertEvalReasonIncidentTransitionFailed, resolveErr)
			}
			incident.Probe = candidate.Incident.Probe
			incident.Check = candidate.Incident.Check
			if notifyErr := s.enqueueNotifications(ctx, flow, candidate.Rule, incident, stoppedEvaluation(candidate.Rule, incident), EventIncidentResolved, now); notifyErr != nil {
				return notifyErr
			}
		}

		if err := s.repo.DeleteInactivePendingEvaluations(ctx, projectID); err != nil {
			return flow.failure(AlertEvalEventLifecycleReconcileFailure, AlertEvalReasonPendingCleanupFailed, err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	flow.success()
	return nil
}

func stoppedEvaluation(rule domainalert.Rule, incident domainalert.Incident) alertcondition.Evaluation {
	summary := alertcondition.MetricSummary{
		Metric:        rule.Condition.Metric,
		MinSamples:    rule.Condition.MinSamples,
		WindowSeconds: rule.Condition.WindowSeconds,
	}
	var stored struct {
		Samples     int64     `json:"samples"`
		WindowStart time.Time `json:"windowStart"`
		WindowEnd   time.Time `json:"windowEnd"`
	}
	if len(incident.LastSummary) > 0 && json.Unmarshal(incident.LastSummary, &stored) == nil {
		summary.Samples = stored.Samples
		summary.WindowStart = stored.WindowStart
		summary.WindowEnd = stored.WindowEnd
	}

	evaluation := alertcondition.Evaluation{
		State:   alertcondition.EvaluationStateNoData,
		Metric:  rule.Condition.Metric,
		Summary: summary,
	}
	if incident.LastValue != nil {
		evaluation.Value = *incident.LastValue
		evaluation.Summary.HasValue = true
	}
	return evaluation
}
