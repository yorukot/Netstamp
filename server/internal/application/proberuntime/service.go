package proberuntime

import (
	"context"
	"errors"
	"time"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

type Service struct {
	probes         ProbeRepository
	results        PingResultRepository
	secretVerifier SecretVerifier
	events         EventRecorder
}

func NewService(probes ProbeRepository, results PingResultRepository, secretVerifier SecretVerifier, events EventRecorder) *Service {
	return &Service{
		probes:         probes,
		results:        results,
		secretVerifier: secretVerifier,
		events:         events,
	}
}

func (s *Service) Hello(ctx context.Context, input RuntimeStatusInput) (HelloOutput, error) {
	ctx, flow := s.startRuntimeFlow(ctx, "probe_runtime.hello", ProbeRuntimeActionHello)
	defer flow.end()

	credential, err := s.authenticate(ctx, flow, input.RuntimeAuthInput)
	if err != nil {
		return HelloOutput{}, flow.authenticationFailure(ProbeRuntimeEventHelloFailure, err)
	}

	statusInput, err := normalizeRuntimeStatus(input, credential.ProbeID)
	if err != nil {
		return HelloOutput{}, flow.businessFailure(ProbeRuntimeEventHelloFailure, ProbeRuntimeReasonInvalidInput, err)
	}
	if _, err := s.probes.UpdateProbeStatus(ctx, statusInput); err != nil {
		return HelloOutput{}, flow.statusUpdateFailure(ProbeRuntimeEventHelloFailure, err)
	}
	flow.success()

	return HelloOutput{
		ServerTime:                    time.Now().UTC(),
		HeartbeatIntervalSeconds:      DefaultHeartbeatIntervalSeconds,
		AssignmentPollIntervalSeconds: DefaultAssignmentPollIntervalSeconds,
	}, nil
}

func (s *Service) Heartbeat(ctx context.Context, input RuntimeStatusInput) (HeartbeatOutput, error) {
	ctx, flow := s.startRuntimeFlow(ctx, "probe_runtime.heartbeat", ProbeRuntimeActionHeartbeat)
	defer flow.end()

	credential, err := s.authenticate(ctx, flow, input.RuntimeAuthInput)
	if err != nil {
		return HeartbeatOutput{}, flow.authenticationFailure(ProbeRuntimeEventHeartbeatFailure, err)
	}

	statusInput, err := normalizeRuntimeStatus(input, credential.ProbeID)
	if err != nil {
		return HeartbeatOutput{}, flow.businessFailure(ProbeRuntimeEventHeartbeatFailure, ProbeRuntimeReasonInvalidInput, err)
	}
	if _, err := s.probes.UpdateProbeStatus(ctx, statusInput); err != nil {
		return HeartbeatOutput{}, flow.statusUpdateFailure(ProbeRuntimeEventHeartbeatFailure, err)
	}
	flow.success()

	return HeartbeatOutput{ServerTime: time.Now().UTC()}, nil
}

func (s *Service) ListAssignments(ctx context.Context, input RuntimeAuthInput) (ListAssignmentsOutput, error) {
	ctx, flow := s.startRuntimeFlow(ctx, "probe_runtime.assignments.list", ProbeRuntimeActionListAssignments)
	defer flow.end()

	credential, err := s.authenticate(ctx, flow, input)
	if err != nil {
		return ListAssignmentsOutput{}, flow.authenticationFailure(ProbeRuntimeEventListAssignmentsFailure, err)
	}

	assignments, err := s.probes.ListAssignments(ctx, credential.ProbeID)
	if err != nil {
		return ListAssignmentsOutput{}, flow.assignmentListFailure(ProbeRuntimeEventListAssignmentsFailure, err)
	}
	flow.success()

	return ListAssignmentsOutput{Assignments: assignments}, nil
}

func (s *Service) SubmitResults(ctx context.Context, input SubmitResultsInput) (SubmitResultsOutput, error) {
	ctx, flow := s.startRuntimeFlow(ctx, "probe_runtime.results.submit", ProbeRuntimeActionSubmitResults)
	defer flow.end()

	credential, err := s.authenticate(ctx, flow, input.RuntimeAuthInput)
	if err != nil {
		return SubmitResultsOutput{}, flow.authenticationFailure(ProbeRuntimeEventSubmitResultsFailure, err)
	}
	resultCount, groups, err := validateResultBatch(input)
	flow.setResultCount(resultCount)
	if err != nil {
		if errors.Is(err, ErrUnsupportedResult) {
			return SubmitResultsOutput{}, flow.businessFailure(ProbeRuntimeEventSubmitResultsFailure, ProbeRuntimeReasonUnsupportedResult, err)
		}
		return SubmitResultsOutput{}, flow.businessFailure(ProbeRuntimeEventSubmitResultsFailure, ProbeRuntimeReasonInvalidInput, err)
	}

	assignments, err := s.probes.ListAssignments(ctx, credential.ProbeID)
	if err != nil {
		return SubmitResultsOutput{}, flow.assignmentListFailure(ProbeRuntimeEventSubmitResultsFailure, err)
	}
	assigned := assignmentsByCheckID(assignments)

	pingResults := make([]domainping.ResultStorageInput, 0, resultCount)
	staleChecks := make([]string, 0)
	latestAssignments := make([]domaincheck.Assignment, 0)
	for _, group := range groups {
		assignment, ok := assigned[group.checkID]
		if !ok || assignment.Type != group.checkType || assignment.Type != domaincheck.TypePing {
			return SubmitResultsOutput{}, flow.businessFailure(ProbeRuntimeEventSubmitResultsFailure, ProbeRuntimeReasonResultConflict, ErrResultConflict)
		}

		if isStaleAssignment(group, assignment) {
			staleChecks = append(staleChecks, group.checkID)
			latestAssignments = append(latestAssignments, assignment)
		}

		for _, result := range group.pingResults {
			storageInput, err := normalizePingResult(result, credential.ProjectID, credential.ProbeID, group.checkID)
			if err != nil {
				return SubmitResultsOutput{}, flow.businessFailure(ProbeRuntimeEventSubmitResultsFailure, ProbeRuntimeReasonInvalidResult, err)
			}
			pingResults = append(pingResults, storageInput)
		}
	}

	if len(pingResults) > 0 {
		if err := s.results.CreatePingResults(ctx, pingResults); err != nil {
			return SubmitResultsOutput{}, flow.resultWriteFailure(err)
		}
	}
	flow.success()

	return SubmitResultsOutput{
		Accepted:     true,
		ResyncNeeded: len(staleChecks) > 0,
		StaleChecks:  staleChecks,
		Assignments:  latestAssignments,
	}, nil
}

func (s *Service) authenticate(ctx context.Context, flow *runtimeFlow, input RuntimeAuthInput) (domainprobe.Credential, error) {
	normalized, err := normalizeRuntimeAuthInput(input)
	if err != nil {
		return domainprobe.Credential{}, err
	}
	flow.setProbeID(normalized.probeID)

	credential, err := s.probes.GetActiveProbeCredential(ctx, normalized.probeID)
	if err != nil {
		return domainprobe.Credential{}, err
	}
	flow.setCredential(credential)
	if !credential.Enabled {
		return domainprobe.Credential{}, ErrProbeDisabled
	}
	if s.secretVerifier == nil {
		return domainprobe.Credential{}, errSecretVerifierMissing
	}
	if !s.secretVerifier.VerifyProbeSecret(normalized.credential, credential.SecretHash) {
		return domainprobe.Credential{}, ErrInvalidCredential
	}

	return credential, nil
}

func assignmentsByCheckID(assignments []domaincheck.Assignment) map[string]domaincheck.Assignment {
	assigned := make(map[string]domaincheck.Assignment, len(assignments))
	for _, assignment := range assignments {
		assigned[assignment.CheckID] = assignment
	}

	return assigned
}

func isStaleAssignment(group normalizedResultGroup, assignment domaincheck.Assignment) bool {
	return group.assignmentID != assignment.ID ||
		group.checkVersion != assignment.CheckVersion ||
		group.selectorVersion != assignment.SelectorVersion
}
