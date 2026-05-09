package proberuntime

import (
	"context"
	"encoding/json"
	"net/netip"
	"strings"
	"time"

	"github.com/google/uuid"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	"github.com/yorukot/netstamp/internal/normalize"
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

func (s *Service) SubmitResults(ctx context.Context, input SubmitResultsInput) error {
	ctx, flow := s.startRuntimeFlow(ctx, "probe_runtime.results.submit", ProbeRuntimeActionSubmitResults)
	defer flow.end()

	credential, err := s.authenticate(ctx, flow, input.RuntimeAuthInput)
	if err != nil {
		return flow.authenticationFailure(ProbeRuntimeEventSubmitResultsFailure, err)
	}
	resultCount := len(input.Ping) + len(input.DNS) + len(input.Traceroute)
	flow.setResultCount(resultCount)

	if resultCount == 0 || resultCount > MaxResultBatchSize {
		return flow.businessFailure(ProbeRuntimeEventSubmitResultsFailure, ProbeRuntimeReasonInvalidInput, ErrInvalidInput)
	}
	if len(input.DNS) > 0 || len(input.Traceroute) > 0 {
		return flow.businessFailure(ProbeRuntimeEventSubmitResultsFailure, ProbeRuntimeReasonUnsupportedResult, ErrUnsupportedResult)
	}

	assignments, err := s.probes.ListAssignments(ctx, credential.ProbeID)
	if err != nil {
		return flow.assignmentListFailure(ProbeRuntimeEventSubmitResultsFailure, err)
	}
	assignedTypes := assignedCheckTypes(assignments)

	pingResults := make([]domainping.ResultStorageInput, 0, len(input.Ping))
	for _, result := range input.Ping {
		storageInput, err := normalizePingResult(result, credential.ProjectID, credential.ProbeID)
		if err != nil {
			return flow.businessFailure(ProbeRuntimeEventSubmitResultsFailure, ProbeRuntimeReasonInvalidResult, err)
		}
		if assignedType, ok := assignedTypes[storageInput.CheckID]; !ok || assignedType != domaincheck.TypePing {
			return flow.businessFailure(ProbeRuntimeEventSubmitResultsFailure, ProbeRuntimeReasonResultConflict, ErrResultConflict)
		}
		pingResults = append(pingResults, storageInput)
	}

	if len(pingResults) > 0 {
		if err := s.results.CreatePingResults(ctx, pingResults); err != nil {
			return flow.resultWriteFailure(err)
		}
	}
	flow.success()

	return nil
}

func (s *Service) authenticate(ctx context.Context, flow *runtimeFlow, input RuntimeAuthInput) (domainprobe.Credential, error) {
	probeID, err := normalizeUUID(input.ProbeID, ErrInvalidInput)
	if err != nil {
		return domainprobe.Credential{}, err
	}
	flow.setProbeID(probeID)
	credentialSecret, err := normalize.RequiredString(input.Credential, ErrInvalidCredential)
	if err != nil {
		return domainprobe.Credential{}, err
	}

	credential, err := s.probes.GetActiveProbeCredential(ctx, probeID)
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
	if !s.secretVerifier.VerifyProbeSecret(credentialSecret, credential.SecretHash) {
		return domainprobe.Credential{}, ErrInvalidCredential
	}

	return credential, nil
}

func normalizeRuntimeStatus(input RuntimeStatusInput, probeID string) (domainprobe.UpdateStatusInput, error) {
	agentVersion, err := normalize.OptionalString(input.AgentVersion, ErrInvalidInput)
	if err != nil {
		return domainprobe.UpdateStatusInput{}, err
	}

	return domainprobe.UpdateStatusInput{
		ProbeID:      probeID,
		State:        domainprobe.StateOnline,
		AgentVersion: agentVersion,
		PublicV4:     input.PublicV4,
		PublicV6:     input.PublicV6,
		Addrs:        append([]netip.Addr(nil), input.Addrs...),
	}, nil
}

func normalizePingResult(input PingResultInput, projectID, probeID string) (domainping.ResultStorageInput, error) {
	checkID, err := normalizeUUID(input.CheckID, ErrInvalidResult)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	status, err := validatePingResultInput(input)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	errorCode, err := normalize.OptionalString(input.ErrorCode, ErrInvalidResult)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	errorMessage, err := normalize.OptionalString(input.ErrorMessage, ErrInvalidResult)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	raw := input.Raw
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	if !json.Valid(raw) {
		return domainping.ResultStorageInput{}, ErrInvalidResult
	}

	return domainping.ResultStorageInput{
		ProjectID:     projectID,
		ProbeID:       probeID,
		CheckID:       checkID,
		StartedAt:     input.StartedAt,
		FinishedAt:    input.FinishedAt,
		DurationMs:    input.DurationMs,
		Status:        status,
		SentCount:     input.SentCount,
		ReceivedCount: input.ReceivedCount,
		LossPercent:   input.LossPercent,
		RttMinMs:      input.RttMinMs,
		RttAvgMs:      input.RttAvgMs,
		RttMedianMs:   input.RttMedianMs,
		RttMaxMs:      input.RttMaxMs,
		RttStddevMs:   input.RttStddevMs,
		RttSamplesMs:  append([]float64(nil), input.RttSamplesMs...),
		ResolvedIP:    input.ResolvedIP,
		IPFamily:      input.IPFamily,
		Raw:           append(json.RawMessage(nil), raw...),
		ErrorCode:     errorCode,
		ErrorMessage:  errorMessage,
	}, nil
}

func validatePingResultInput(input PingResultInput) (domainping.Status, error) {
	status, err := parsePingStatus(input.Status)
	if err != nil {
		return "", err
	}
	if invalidPingResultTiming(input) || invalidPingResultCounts(input) || invalidRTTs(input) {
		return "", ErrInvalidResult
	}

	return status, nil
}

func parsePingStatus(value string) (domainping.Status, error) {
	switch domainping.Status(strings.TrimSpace(value)) {
	case domainping.StatusSuccessful:
		return domainping.StatusSuccessful, nil
	case domainping.StatusTimeout:
		return domainping.StatusTimeout, nil
	case domainping.StatusError:
		return domainping.StatusError, nil
	default:
		return "", ErrInvalidResult
	}
}

func invalidPingResultTiming(input PingResultInput) bool {
	return input.StartedAt.IsZero() ||
		input.FinishedAt.IsZero() ||
		input.FinishedAt.Before(input.StartedAt) ||
		input.DurationMs < 0
}

func invalidPingResultCounts(input PingResultInput) bool {
	return input.SentCount < 0 ||
		input.ReceivedCount < 0 ||
		input.ReceivedCount > input.SentCount ||
		input.LossPercent < 0 ||
		input.LossPercent > 100
}

func invalidRTTs(input PingResultInput) bool {
	return hasNegativeRTTMetric(input) || invalidRTTOrder(input) || hasNegativeRTTSample(input.RttSamplesMs)
}

func hasNegativeRTTMetric(input PingResultInput) bool {
	values := []*float64{
		input.RttMinMs,
		input.RttAvgMs,
		input.RttMedianMs,
		input.RttMaxMs,
		input.RttStddevMs,
	}
	for _, value := range values {
		if value != nil && *value < 0 {
			return true
		}
	}

	return false
}

func invalidRTTOrder(input PingResultInput) bool {
	return greaterThan(input.RttMinMs, input.RttMaxMs) ||
		greaterThan(input.RttMinMs, input.RttAvgMs) ||
		greaterThan(input.RttAvgMs, input.RttMaxMs) ||
		greaterThan(input.RttMinMs, input.RttMedianMs) ||
		greaterThan(input.RttMedianMs, input.RttMaxMs)
}

func hasNegativeRTTSample(samples []float64) bool {
	for _, sample := range samples {
		if sample < 0 {
			return true
		}
	}

	return false
}

func greaterThan(left, right *float64) bool {
	return left != nil && right != nil && *left > *right
}

func normalizeUUID(value string, invalidErr error) (string, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return "", invalidErr
	}

	return parsed.String(), nil
}

func assignedCheckTypes(assignments []domaincheck.Assignment) map[string]domaincheck.Type {
	assigned := make(map[string]domaincheck.Type, len(assignments))
	for _, assignment := range assignments {
		assigned[assignment.CheckID] = assignment.Type
	}

	return assigned
}
