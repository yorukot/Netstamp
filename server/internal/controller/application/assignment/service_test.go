package assignment

import (
	"context"
	"errors"
	"testing"
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
)

const (
	testProjectID = "11111111-1111-1111-1111-111111111111"
	testProbeID   = "22222222-2222-2222-2222-222222222222"
	testCheckID   = "33333333-3333-3333-3333-333333333333"
	testLabelID   = "44444444-4444-4444-4444-444444444444"
)

func TestRefreshProbeCheckAssignmentsForProbeCallsRepository(t *testing.T) {
	repo := &recordingRepository{}
	events := &recordingEventRecorder{}
	service := NewService(repo, nil, events)

	if err := service.RefreshProbeCheckAssignmentsForProbe(context.Background(), testProjectID, testProbeID); err != nil {
		t.Fatalf("expected refresh to succeed: %v", err)
	}

	if repo.method != "refresh_probe" || repo.projectID != testProjectID || repo.probeID != testProbeID {
		t.Fatalf("unexpected repository call: %#v", repo)
	}
	assertEnqueued(t, repo, domainassignment.RefreshTarget{ProjectID: testProjectID, Type: domainassignment.RefreshTargetProbe, TargetID: testProbeID})
	assertLastEvent(t, events, AssignmentEventRefreshProbeSuccess, AssignmentOutcomeSuccess, "")
}

func TestRefreshProbeCheckAssignmentsForProjectCallsRepository(t *testing.T) {
	repo := &recordingRepository{}
	events := &recordingEventRecorder{}
	service := NewService(repo, nil, events)

	if err := service.RefreshProbeCheckAssignmentsForProject(context.Background(), testProjectID); err != nil {
		t.Fatalf("expected refresh to succeed: %v", err)
	}

	if repo.method != "refresh_project" || repo.projectID != testProjectID {
		t.Fatalf("unexpected repository call: %#v", repo)
	}
	assertEnqueued(t, repo, domainassignment.RefreshTarget{ProjectID: testProjectID, Type: domainassignment.RefreshTargetProject, TargetID: testProjectID})
	assertLastEvent(t, events, AssignmentEventRefreshProjectSuccess, AssignmentOutcomeSuccess, "")
}

func TestRefreshProbeCheckAssignmentsForCheckCallsRepository(t *testing.T) {
	repo := &recordingRepository{}
	events := &recordingEventRecorder{}
	service := NewService(repo, nil, events)

	if err := service.RefreshProbeCheckAssignmentsForCheck(context.Background(), testProjectID, testCheckID); err != nil {
		t.Fatalf("expected refresh to succeed: %v", err)
	}

	if repo.method != "refresh_check" || repo.projectID != testProjectID || repo.checkID != testCheckID {
		t.Fatalf("unexpected repository call: %#v", repo)
	}
	assertEnqueued(t, repo, domainassignment.RefreshTarget{ProjectID: testProjectID, Type: domainassignment.RefreshTargetCheck, TargetID: testCheckID})
	assertLastEvent(t, events, AssignmentEventRefreshCheckSuccess, AssignmentOutcomeSuccess, "")
}

func TestRefreshProbeCheckAssignmentsForLabelCallsRepository(t *testing.T) {
	repo := &recordingRepository{}
	events := &recordingEventRecorder{}
	service := NewService(repo, nil, events)

	if err := service.RefreshProbeCheckAssignmentsForLabel(context.Background(), testProjectID, testLabelID); err != nil {
		t.Fatalf("expected refresh to succeed: %v", err)
	}

	if repo.method != "refresh_label" || repo.projectID != testProjectID || repo.labelID != testLabelID {
		t.Fatalf("unexpected repository call: %#v", repo)
	}
	assertEnqueued(t, repo, domainassignment.RefreshTarget{ProjectID: testProjectID, Type: domainassignment.RefreshTargetLabel, TargetID: testLabelID})
	assertLastEvent(t, events, AssignmentEventRefreshLabelSuccess, AssignmentOutcomeSuccess, "")
}

func TestDeleteProbeCheckAssignmentsForProbeCallsRepository(t *testing.T) {
	repo := &recordingRepository{}
	events := &recordingEventRecorder{}
	service := NewService(repo, nil, events)

	if err := service.DeleteProbeCheckAssignmentsForProbe(context.Background(), testProjectID, testProbeID); err != nil {
		t.Fatalf("expected delete to succeed: %v", err)
	}

	if repo.method != "delete_probe" || repo.projectID != testProjectID || repo.probeID != testProbeID {
		t.Fatalf("unexpected repository call: %#v", repo)
	}
	assertEnqueued(t, repo, domainassignment.RefreshTarget{ProjectID: testProjectID, Type: domainassignment.RefreshTargetProbe, TargetID: testProbeID})
	assertLastEvent(t, events, AssignmentEventDeleteProbeSuccess, AssignmentOutcomeSuccess, "")
}

func TestDeleteProbeCheckAssignmentsForCheckCallsRepository(t *testing.T) {
	repo := &recordingRepository{}
	events := &recordingEventRecorder{}
	service := NewService(repo, nil, events)

	if err := service.DeleteProbeCheckAssignmentsForCheck(context.Background(), testProjectID, testCheckID); err != nil {
		t.Fatalf("expected delete to succeed: %v", err)
	}

	if repo.method != "delete_check" || repo.projectID != testProjectID || repo.checkID != testCheckID {
		t.Fatalf("unexpected repository call: %#v", repo)
	}
	assertEnqueued(t, repo, domainassignment.RefreshTarget{ProjectID: testProjectID, Type: domainassignment.RefreshTargetCheck, TargetID: testCheckID})
	assertLastEvent(t, events, AssignmentEventDeleteCheckSuccess, AssignmentOutcomeSuccess, "")
}

func TestRefreshRejectsInvalidInputBeforeRepository(t *testing.T) {
	repo := &recordingRepository{}
	events := &recordingEventRecorder{}
	service := NewService(repo, nil, events)

	err := service.RefreshProbeCheckAssignmentsForProbe(context.Background(), "not-a-uuid", testProbeID)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if repo.method != "" {
		t.Fatalf("expected repository not to be called, got %#v", repo)
	}
	if len(repo.enqueued) != 0 {
		t.Fatalf("expected no refresh job to be enqueued, got %#v", repo.enqueued)
	}
	assertLastEvent(t, events, AssignmentEventRefreshProbeFailure, AssignmentOutcomeFailure, AssignmentReasonInvalidInput)
}

func TestRefreshRecordsRepositoryFailure(t *testing.T) {
	wantErr := errors.New("refresh failed")
	repo := &recordingRepository{err: wantErr}
	events := &recordingEventRecorder{}
	service := NewService(repo, nil, events)

	err := service.RefreshProbeCheckAssignmentsForLabel(context.Background(), testProjectID, testLabelID)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
	assertLastEvent(t, events, AssignmentEventRefreshLabelFailure, AssignmentOutcomeFailure, AssignmentReasonRefreshFailed)
}

func TestRefreshRecordsEnqueueFailureBeforeSyncRefresh(t *testing.T) {
	wantErr := errors.New("enqueue failed")
	repo := &recordingRepository{enqueueErr: wantErr}
	events := &recordingEventRecorder{}
	service := NewService(repo, nil, events)

	err := service.RefreshProbeCheckAssignmentsForLabel(context.Background(), testProjectID, testLabelID)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected enqueue error, got %v", err)
	}
	if repo.method != "" {
		t.Fatalf("expected sync refresh not to run, got %#v", repo)
	}
	assertLastEvent(t, events, AssignmentEventRefreshLabelFailure, AssignmentOutcomeFailure, AssignmentReasonRefreshFailed)
}

func TestWorkerRefreshRunnerDoesNotEnqueueRefreshJob(t *testing.T) {
	repo := &recordingRepository{}
	events := &recordingEventRecorder{}
	service := NewService(repo, nil, events)
	runner := NewWorkerRefreshRunner(service)

	if err := runner.RefreshProbeCheckAssignmentsForProbe(context.Background(), testProjectID, testProbeID); err != nil {
		t.Fatalf("expected refresh to succeed: %v", err)
	}

	if repo.method != "refresh_probe" || repo.projectID != testProjectID || repo.probeID != testProbeID {
		t.Fatalf("unexpected repository call: %#v", repo)
	}
	if len(repo.enqueued) != 0 {
		t.Fatalf("expected worker runner not to enqueue refresh jobs, got %#v", repo.enqueued)
	}
	assertLastEvent(t, events, AssignmentEventRefreshProbeSuccess, AssignmentOutcomeSuccess, "")
}

func assertLastEvent(t *testing.T, recorder *recordingEventRecorder, name AssignmentEventName, outcome AssignmentEventOutcome, reason AssignmentEventReason) {
	t.Helper()

	if len(recorder.events) == 0 {
		t.Fatal("expected an assignment event")
	}
	got := recorder.events[len(recorder.events)-1]
	if got.Name != name || got.Outcome != outcome || got.Reason != reason {
		t.Fatalf("unexpected event:\n got: %#v\nwant name=%q outcome=%q reason=%q", got, name, outcome, reason)
	}
}

func assertEnqueued(t *testing.T, repo *recordingRepository, want domainassignment.RefreshTarget) {
	t.Helper()

	if len(repo.enqueued) != 1 {
		t.Fatalf("expected one refresh job to be enqueued, got %#v", repo.enqueued)
	}
	if repo.enqueued[0] != want {
		t.Fatalf("unexpected refresh job target:\n got: %#v\nwant: %#v", repo.enqueued[0], want)
	}
	if repo.maxAttempts != domainassignment.DefaultRefreshJobMaxAttempts {
		t.Fatalf("expected default max attempts, got %d", repo.maxAttempts)
	}
}

type recordingRepository struct {
	method      string
	projectID   string
	probeID     string
	checkID     string
	labelID     string
	err         error
	enqueueErr  error
	enqueued    []domainassignment.RefreshTarget
	maxAttempts int32
}

func (r *recordingRepository) EnqueueRefreshJob(_ context.Context, target domainassignment.RefreshTarget, maxAttempts int32) error {
	r.enqueued = append(r.enqueued, target)
	r.maxAttempts = maxAttempts
	return r.enqueueErr
}

func (r *recordingRepository) ClaimRefreshJobs(_ context.Context, _ int32, _ time.Time) ([]domainassignment.RefreshJob, error) {
	return nil, nil
}

func (r *recordingRepository) MarkRefreshJobSucceeded(_ context.Context, _ string, _ time.Time) error {
	return nil
}

func (r *recordingRepository) MarkRefreshJobRetry(_ context.Context, _ string, _ time.Time, _, _, _ string) error {
	return nil
}

func (r *recordingRepository) MarkRefreshJobFailed(_ context.Context, _, _, _, _ string) error {
	return nil
}

func (r *recordingRepository) MarkRefreshJobDiscarded(_ context.Context, _, _, _, _ string) error {
	return nil
}

func (r *recordingRepository) ListProbeRefreshCandidatesForProject(_ context.Context, projectID string) ([]ProbeAssignmentCandidate, error) {
	r.method = "refresh_project"
	r.projectID = projectID
	return nil, r.err
}

func (r *recordingRepository) GetProbeRefreshCandidate(_ context.Context, projectID, probeID string) (ProbeAssignmentCandidate, error) {
	r.method = "refresh_probe"
	r.projectID = projectID
	r.probeID = probeID
	return ProbeAssignmentCandidate{ProjectID: projectID, ProbeID: probeID, Enabled: true}, r.err
}

func (r *recordingRepository) ListProbeRefreshCandidatesForLabel(_ context.Context, projectID, labelID string) ([]ProbeAssignmentCandidate, error) {
	r.method = "refresh_label"
	r.projectID = projectID
	r.labelID = labelID
	return nil, r.err
}

func (r *recordingRepository) ListCheckRefreshCandidatesForProject(_ context.Context, projectID string) ([]CheckAssignmentCandidate, error) {
	r.projectID = projectID
	return nil, nil
}

func (r *recordingRepository) GetCheckRefreshCandidate(_ context.Context, projectID, checkID string) (CheckAssignmentCandidate, error) {
	r.method = "refresh_check"
	r.projectID = projectID
	r.checkID = checkID
	selector, err := domainselector.Parse(nil)
	if err != nil {
		return CheckAssignmentCandidate{}, err
	}
	return CheckAssignmentCandidate{
		Check:           domaincheck.Check{ProjectID: projectID, ID: checkID},
		Selector:        selector,
		SelectorVersion: "selector-version",
		CheckVersion:    "check-version",
	}, r.err
}

func (r *recordingRepository) ListSelectorPreviewCandidates(_ context.Context, projectID string) ([]ProbeAssignmentCandidate, error) {
	if r.method == "" {
		r.method = "preview"
	}
	r.projectID = projectID
	return nil, r.err
}

func (r *recordingRepository) UpsertProbeCheckAssignment(_ context.Context, _ AssignmentWrite) error {
	return r.err
}

func (r *recordingRepository) DeleteStaleAssignmentsForProbe(_ context.Context, projectID, probeID string, _ []string) error {
	r.projectID = projectID
	r.probeID = probeID
	return r.err
}

func (r *recordingRepository) DeleteStaleAssignmentsForCheck(_ context.Context, projectID, checkID, _, _ string, _ []string) error {
	r.projectID = projectID
	r.checkID = checkID
	return r.err
}

func (r *recordingRepository) DeleteProbeCheckAssignmentsForProbe(_ context.Context, projectID, probeID string) error {
	r.method = "delete_probe"
	r.projectID = projectID
	r.probeID = probeID
	return r.err
}

func (r *recordingRepository) DeleteProbeCheckAssignmentsForCheck(_ context.Context, projectID, checkID string) error {
	r.method = "delete_check"
	r.projectID = projectID
	r.checkID = checkID
	return r.err
}

func (r *recordingRepository) ListProjectAssignments(_ context.Context, input domainassignment.Query) ([]domainassignment.Assignment, error) {
	r.method = "list"
	r.projectID = input.ProjectID
	r.probeID = input.ProbeID
	r.checkID = input.CheckID
	return nil, r.err
}

type recordingEventRecorder struct {
	events []AssignmentEvent
}

func (r *recordingEventRecorder) RecordAssignmentEvent(_ context.Context, event AssignmentEvent) {
	r.events = append(r.events, event)
}
