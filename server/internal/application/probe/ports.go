package probe

import (
	"context"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

type Repository interface {
	GetProjectIDForUser(ctx context.Context, projectRef string, userID string) (string, error)
	CreateProbe(ctx context.Context, input domainprobe.CreateProbeStorageInput) (domainprobe.Probe, error)
}

type SecretGenerator interface {
	GenerateProbeSecret() (plaintext string, hash string, err error)
}

type EventRecorder interface {
	RecordProbeEvent(ctx context.Context, event ProbeEvent)
}

type ProbeEventName string

const (
	ProbeEventCreateSuccess ProbeEventName = "probe.create.success"
	ProbeEventCreateFailure ProbeEventName = "probe.create.failure"
)

type ProbeEventAction string

const (
	ProbeActionCreate ProbeEventAction = "create"
)

type ProbeEventOutcome string

const (
	ProbeOutcomeSuccess ProbeEventOutcome = "success"
	ProbeOutcomeFailure ProbeEventOutcome = "failure"
)

type ProbeEventReason string

const (
	ProbeReasonInvalidInput           ProbeEventReason = "invalid_input"
	ProbeReasonProjectNotFound        ProbeEventReason = "project_not_found"
	ProbeReasonLabelNotFound          ProbeEventReason = "label_not_found"
	ProbeReasonProjectLookupFailed    ProbeEventReason = "project_lookup_failed"
	ProbeReasonSecretGeneratorMissing ProbeEventReason = "secret_generator_missing"
	ProbeReasonSecretGenerateFailed   ProbeEventReason = "secret_generate_failed"
	ProbeReasonProbeCreateFailed      ProbeEventReason = "probe_create_failed"
)

type ProbeEvent struct {
	Name        ProbeEventName
	Action      ProbeEventAction
	Outcome     ProbeEventOutcome
	Reason      ProbeEventReason
	ActorUserID string
	ProjectID   string
	ProjectRef  string
	ProbeID     string
	Err         error
}
