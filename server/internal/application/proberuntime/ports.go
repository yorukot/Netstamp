package proberuntime

import (
	"context"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

type ProbeRepository interface {
	GetActiveProbeCredential(ctx context.Context, probeID string) (domainprobe.Credential, error)
	UpdateProbeStatus(ctx context.Context, input domainprobe.UpdateStatusInput) (domainprobe.Status, error)
	ListAssignments(ctx context.Context, probeID string) ([]domaincheck.Assignment, error)
}

type PingResultRepository interface {
	CreatePingResults(ctx context.Context, inputs []domainping.ResultStorageInput) error
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
	ProbeRuntimeEventListAssignmentsFailure ProbeRuntimeEventName = "probe_runtime.assignments.list.failure"
	ProbeRuntimeEventSubmitResultsFailure   ProbeRuntimeEventName = "probe_runtime.results.submit.failure"
)

type ProbeRuntimeEventAction string

const (
	ProbeRuntimeActionHello           ProbeRuntimeEventAction = "hello"
	ProbeRuntimeActionHeartbeat       ProbeRuntimeEventAction = "heartbeat"
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
	ProbeRuntimeReasonInvalidResult         ProbeRuntimeEventReason = "invalid_result"
	ProbeRuntimeReasonResultConflict        ProbeRuntimeEventReason = "result_conflict"
	ProbeRuntimeReasonUnsupportedResult     ProbeRuntimeEventReason = "unsupported_result"
	ProbeRuntimeReasonCredentialLookupFail  ProbeRuntimeEventReason = "credential_lookup_failed" //nolint:gosec // Event reason label, not a credential.
	ProbeRuntimeReasonSecretVerifierMissing ProbeRuntimeEventReason = "secret_verifier_missing"  //nolint:gosec // Event reason label, not a credential.
	ProbeRuntimeReasonStatusUpdateFailed    ProbeRuntimeEventReason = "status_update_failed"
	ProbeRuntimeReasonAssignmentListFailed  ProbeRuntimeEventReason = "assignment_list_failed"
	ProbeRuntimeReasonResultWriteFailed     ProbeRuntimeEventReason = "result_write_failed"
)

type ProbeRuntimeEvent struct {
	Name        ProbeRuntimeEventName
	Action      ProbeRuntimeEventAction
	Outcome     ProbeRuntimeEventOutcome
	Reason      ProbeRuntimeEventReason
	ProbeID     string
	ProjectID   string
	ResultCount *int
	Err         error
}
