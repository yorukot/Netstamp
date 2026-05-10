package probe

import (
	"context"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type Repository interface {
	CreateProbe(ctx context.Context, input domainprobe.CreateProbeStorageInput) (domainprobe.Probe, error)
	ListProbesForProject(ctx context.Context, projectID string) ([]domainprobe.Probe, error)
	GetProbeForProject(ctx context.Context, projectID, probeID string) (domainprobe.Probe, error)
	UpdateProbe(ctx context.Context, input domainprobe.UpdateProbeStorageInput) (domainprobe.Probe, error)
	SoftDeleteProbe(ctx context.Context, projectID, probeID string) error
	RotateProbeSecret(ctx context.Context, input domainprobe.RotateProbeSecretStorageInput) error
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
	ProbeEventCreateFailure       ProbeEventName = "probe.create.failure"
	ProbeEventUpdateSuccess       ProbeEventName = "probe.update.success"
	ProbeEventUpdateFailure       ProbeEventName = "probe.update.failure"
	ProbeEventDeleteSuccess       ProbeEventName = "probe.delete.success"
	ProbeEventDeleteFailure       ProbeEventName = "probe.delete.failure"
	ProbeEventSecretRotateSuccess ProbeEventName = "probe.secret.rotate.success"
	ProbeEventSecretRotateFailure ProbeEventName = "probe.secret.rotate.failure"
	ProbeEventListFailure         ProbeEventName = "probe.list.failure"
	ProbeEventGetFailure          ProbeEventName = "probe.get.failure"
)

type ProbeEventAction string

const (
	ProbeActionCreate       ProbeEventAction = "create"
	ProbeActionUpdate       ProbeEventAction = "update"
	ProbeActionDelete       ProbeEventAction = "delete"
	ProbeActionSecretRotate ProbeEventAction = "secret_rotate"
	ProbeActionList         ProbeEventAction = "list"
	ProbeActionGet          ProbeEventAction = "get"
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
	ProbeReasonProbeNotFound          ProbeEventReason = "probe_not_found"
	ProbeReasonForbidden              ProbeEventReason = "forbidden"
	ProbeReasonLabelNotFound          ProbeEventReason = "label_not_found"
	ProbeReasonProjectLookupFailed    ProbeEventReason = "project_lookup_failed"
	ProbeReasonRoleLookupFailed       ProbeEventReason = "role_lookup_failed"
	ProbeReasonLabelLookupFailed      ProbeEventReason = "label_lookup_failed"
	ProbeReasonProbeLookupFailed      ProbeEventReason = "probe_lookup_failed"
	ProbeReasonSecretGeneratorMissing ProbeEventReason = "secret_generator_missing"
	ProbeReasonSecretGenerateFailed   ProbeEventReason = "secret_generate_failed"
	ProbeReasonProbeCreateFailed      ProbeEventReason = "probe_create_failed"
	ProbeReasonProbeListFailed        ProbeEventReason = "probe_list_failed"
	ProbeReasonProbeUpdateFailed      ProbeEventReason = "probe_update_failed"
	ProbeReasonProbeDeleteFailed      ProbeEventReason = "probe_delete_failed"
	ProbeReasonSecretRotateFailed     ProbeEventReason = "secret_rotate_failed"
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
