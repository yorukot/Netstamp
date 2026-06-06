package proberuntime

import (
	"context"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type ProbeRepository interface {
	GetActiveProbeCredential(ctx context.Context, probeID string) (domainprobe.Credential, error)
	UpdateProbeStatus(ctx context.Context, input domainprobe.Status) (domainprobe.Status, error)
	UpdateProbeIPFamilyCapabilities(ctx context.Context, input domainprobe.IPFamilyCapabilities) (domainprobe.Status, error)
	ListAssignments(ctx context.Context, probeID string) ([]domainassignment.Assignment, error)
	ListActiveAssignmentsForProbeChecks(ctx context.Context, probeID string, checkIDs []string) ([]domainassignment.Assignment, error)
}

type PingResultRepository interface {
	CreatePingResults(ctx context.Context, inputs []domainping.ResultStorageInput) error
}

type TCPResultRepository interface {
	CreateTCPResults(ctx context.Context, inputs []domaintcp.ResultStorageInput) error
}

type TracerouteResultRepository interface {
	CreateTracerouteResults(ctx context.Context, inputs []domaintraceroute.ResultStorageInput) error
}

type SecretVerifier interface {
	VerifyProbeSecret(secret, expectedHash string) bool
}

type EventRecorder interface {
	RecordProbeRuntimeEvent(ctx context.Context, event ProbeRuntimeEvent)
}

type ProbeRuntimeEventName string

const (
	ProbeRuntimeEventHelloFailure           ProbeRuntimeEventName = "probe_runtime.hello.failure"
	ProbeRuntimeEventHeartbeatFailure       ProbeRuntimeEventName = "probe_runtime.heartbeat.failure"
	ProbeRuntimeEventIPFamilyUpdateFailure  ProbeRuntimeEventName = "probe_runtime.ip_family_capabilities.update.failure"
	ProbeRuntimeEventListAssignmentsFailure ProbeRuntimeEventName = "probe_runtime.assignments.list.failure"
	ProbeRuntimeEventSubmitResultsFailure   ProbeRuntimeEventName = "probe_runtime.results.submit.failure"
)

type ProbeRuntimeEventAction string

const (
	ProbeRuntimeActionHello           ProbeRuntimeEventAction = "hello"
	ProbeRuntimeActionHeartbeat       ProbeRuntimeEventAction = "heartbeat"
	ProbeRuntimeActionUpdateIPFamily  ProbeRuntimeEventAction = "ip_family_capabilities.update"
	ProbeRuntimeActionListAssignments ProbeRuntimeEventAction = "assignments.list"
	ProbeRuntimeActionSubmitResults   ProbeRuntimeEventAction = "results.submit"
)

type ProbeRuntimeEventOutcome string

const (
	ProbeRuntimeOutcomeSuccess ProbeRuntimeEventOutcome = "success"
	ProbeRuntimeOutcomeFailure ProbeRuntimeEventOutcome = "failure"
)

type ProbeRuntimeEventReason string

const (
	ProbeRuntimeReasonInvalidInput          ProbeRuntimeEventReason = "invalid_input"
	ProbeRuntimeReasonInvalidCredential     ProbeRuntimeEventReason = "invalid_credential" //nolint:gosec // Event reason label, not a credential.
	ProbeRuntimeReasonProbeNotFound         ProbeRuntimeEventReason = "probe_not_found"
	ProbeRuntimeReasonProbeDisabled         ProbeRuntimeEventReason = "probe_disabled"
	ProbeRuntimeReasonCredentialLookupFail  ProbeRuntimeEventReason = "credential_lookup_failed" //nolint:gosec // Event reason label, not a credential.
	ProbeRuntimeReasonSecretVerifierMissing ProbeRuntimeEventReason = "secret_verifier_missing"  //nolint:gosec // Event reason label, not a credential.
	ProbeRuntimeReasonStatusUpdateFailed    ProbeRuntimeEventReason = "status_update_failed"
	ProbeRuntimeReasonIPFamilyUpdateFailed  ProbeRuntimeEventReason = "ip_family_capability_update_failed"
	ProbeRuntimeReasonAssignmentListFailed  ProbeRuntimeEventReason = "assignment_list_failed"
	ProbeRuntimeReasonAssignmentLookupFail  ProbeRuntimeEventReason = "assignment_lookup_failed"
	ProbeRuntimeReasonResultWriteFailed     ProbeRuntimeEventReason = "result_write_failed"
)

type ProbeRuntimeEvent struct {
	Name      ProbeRuntimeEventName
	Action    ProbeRuntimeEventAction
	Outcome   ProbeRuntimeEventOutcome
	Reason    ProbeRuntimeEventReason
	ProbeID   string
	ProjectID string
	Err       error
}
