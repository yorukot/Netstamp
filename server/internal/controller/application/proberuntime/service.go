package proberuntime

import (
	"context"
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

type Service struct {
	probes         ProbeRepository
	pings          PingResultRepository
	secretVerifier SecretVerifier
	events         EventRecorder
}

func NewService(probes ProbeRepository, pings PingResultRepository, secretVerifier SecretVerifier, events EventRecorder) *Service {
	return &Service{
		probes:         probes,
		pings:          pings,
		secretVerifier: secretVerifier,
		events:         events,
	}
}

func (s *Service) Hello(ctx context.Context, input RuntimeAuthInput) (HelloOutput, error) {
	ctx, flow := s.startRuntimeFlow(ctx, "probe_runtime.hello", ProbeRuntimeActionHello)
	defer flow.end()

	if _, err := s.authenticate(ctx, flow, input); err != nil {
		return HelloOutput{}, flow.authenticationFailure(ProbeRuntimeEventHelloFailure, err)
	}
	flow.success()

	return HelloOutput{
		ServerTime:                   time.Now().UTC(),
		MinimumSupportedAgentVersion: domainprobe.DefaultMinimumSupportedAgentVersion,
		Config:                       domainprobe.DefaultRuntimeConfig(),
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

	return ListAssignmentsOutput{
		ServerTime:  time.Now().UTC(),
		Config:      domainprobe.DefaultRuntimeConfig(),
		Assignments: assignments,
	}, nil
}

func (s *Service) SubmitResults(ctx context.Context, input SubmitResultsInput) (SubmitResultsOutput, error) {
	ctx, flow := s.startRuntimeFlow(ctx, "probe_runtime.results.submit", ProbeRuntimeActionSubmitResults)
	defer flow.end()

	credential, err := s.authenticate(ctx, flow, input.RuntimeAuthInput)
	if err != nil {
		return SubmitResultsOutput{}, flow.authenticationFailure(ProbeRuntimeEventSubmitResultsFailure, err)
	}

	normalized, err := normalizeSubmitResults(input)
	if err != nil {
		return SubmitResultsOutput{}, flow.businessFailure(ProbeRuntimeEventSubmitResultsFailure, ProbeRuntimeReasonInvalidInput, err)
	}

	assignments, err := s.probes.ListActiveAssignmentsForProbeChecks(ctx, credential.ProbeID, normalized.checkIDs)
	if err != nil {
		return SubmitResultsOutput{}, flow.assignmentLookupFailure(ProbeRuntimeEventSubmitResultsFailure, err)
	}
	assignmentByCheckID := assignmentsByCheckID(assignments)

	pingResults := make([]domainping.ResultStorageInput, 0, normalized.accepted)
	for _, group := range normalized.groups {
		assignment, ok := assignmentByCheckID[group.checkID]
		if !ok {
			return SubmitResultsOutput{}, flow.businessFailure(ProbeRuntimeEventSubmitResultsFailure, ProbeRuntimeReasonInvalidInput, invalidRuntimeField(resultGroupField(group.index, "checkId"), "check is not an active assignment for this probe", group.checkID))
		}
		if assignment.Check == nil || assignment.Check.Type != group.checkType {
			return SubmitResultsOutput{}, flow.businessFailure(ProbeRuntimeEventSubmitResultsFailure, ProbeRuntimeReasonInvalidInput, invalidRuntimeField(resultGroupField(group.index, "type"), "does not match assigned check type", string(group.checkType)))
		}

		switch group.checkType {
		case domaincheck.TypePing:
			for _, result := range group.ping {
				result.ProjectID = assignment.ProjectID
				result.ProbeID = credential.ProbeID
				result.CheckID = assignment.CheckID
				pingResults = append(pingResults, result)
			}
		default:
			return SubmitResultsOutput{}, flow.businessFailure(ProbeRuntimeEventSubmitResultsFailure, ProbeRuntimeReasonInvalidInput, invalidRuntimeField(resultGroupField(group.index, "type"), "unsupported result type", string(group.checkType)))
		}
	}

	if len(pingResults) > 0 {
		if s.pings == nil {
			return SubmitResultsOutput{}, flow.resultWriteFailure(ProbeRuntimeEventSubmitResultsFailure, errPingRepositoryMissing)
		}
		if err := s.pings.CreatePingResults(ctx, pingResults); err != nil {
			return SubmitResultsOutput{}, flow.resultWriteFailure(ProbeRuntimeEventSubmitResultsFailure, err)
		}
	}
	flow.success()

	return SubmitResultsOutput{Accepted: normalized.accepted, ServerTime: time.Now().UTC()}, nil
}

func assignmentsByCheckID(assignments []domainassignment.Assignment) map[string]domainassignment.Assignment {
	byCheckID := make(map[string]domainassignment.Assignment, len(assignments))
	for _, assignment := range assignments {
		byCheckID[assignment.CheckID] = assignment
	}

	return byCheckID
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
		return domainprobe.Credential{}, domainprobe.ErrProbeDisabled
	}
	if s.secretVerifier == nil {
		return domainprobe.Credential{}, errSecretVerifierMissing
	}
	if !s.secretVerifier.VerifyProbeSecret(normalized.credential, credential.SecretHash) {
		return domainprobe.Credential{}, domainprobe.ErrInvalidCredential
	}

	return credential, nil
}
