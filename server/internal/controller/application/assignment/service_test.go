package assignment

import (
	"context"
	"errors"
	"testing"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
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

type recordingRepository struct {
	method    string
	projectID string
	probeID   string
	checkID   string
	labelID   string
	err       error
}

func (r *recordingRepository) RefreshProbeCheckAssignmentsForProbe(_ context.Context, projectID, probeID string) error {
	r.method = "refresh_probe"
	r.projectID = projectID
	r.probeID = probeID
	return r.err
}

func (r *recordingRepository) RefreshProbeCheckAssignmentsForProject(_ context.Context, projectID string) error {
	r.method = "refresh_project"
	r.projectID = projectID
	return r.err
}

func (r *recordingRepository) RefreshProbeCheckAssignmentsForCheck(_ context.Context, projectID, checkID string) error {
	r.method = "refresh_check"
	r.projectID = projectID
	r.checkID = checkID
	return r.err
}

func (r *recordingRepository) RefreshProbeCheckAssignmentsForLabel(_ context.Context, projectID, labelID string) error {
	r.method = "refresh_label"
	r.projectID = projectID
	r.labelID = labelID
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

func (r *recordingRepository) ListSelectorPreviewProbes(_ context.Context, projectID string, _ domainselector.Selector) ([]domainprobe.Probe, error) {
	r.method = "preview"
	r.projectID = projectID
	return nil, r.err
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
