package probe

import (
	"context"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Repository interface {
	CreateProbe(ctx context.Context, input domainprobe.CreateProbeStorageInput) (domainprobe.Probe, error)
}

type ProjectAccess interface {
	GetProjectForUser(ctx context.Context, projectRef, userID string) (domainproject.Project, error)
	GetMemberRole(ctx context.Context, projectID, userID string) (domainproject.Role, error)
}

type LabelAccess interface {
	GetActiveLabelsByIDsForProject(ctx context.Context, projectID string, labelIDs []string) ([]domainlabel.Label, error)
}

type SecretGenerator interface {
	GenerateProbeSecret() (plaintext, hash string, err error)
}

type EventRecorder interface {
	RecordProbeEvent(ctx context.Context, event ProbeEvent)
}

type ProbeEventName string

const (
	ProbeEventCreateFailure ProbeEventName = "probe.create.failure"
)

type ProbeEventAction string

const (
	ProbeActionCreate ProbeEventAction = "create"
)

type ProbeEventOutcome string

const (
	ProbeOutcomeFailure ProbeEventOutcome = "failure"
)

type ProbeEventReason string

const (
	ProbeReasonInvalidInput           ProbeEventReason = "invalid_input"
	ProbeReasonProjectNotFound        ProbeEventReason = "project_not_found"
	ProbeReasonForbidden              ProbeEventReason = "forbidden"
	ProbeReasonLabelNotFound          ProbeEventReason = "label_not_found"
	ProbeReasonProjectLookupFailed    ProbeEventReason = "project_lookup_failed"
	ProbeReasonRoleLookupFailed       ProbeEventReason = "role_lookup_failed"
	ProbeReasonLabelLookupFailed      ProbeEventReason = "label_lookup_failed"
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
