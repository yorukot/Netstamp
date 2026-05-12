package proberuntime

import (
	"context"
	"errors"
	"testing"
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

const (
	testProbeID      = "33333333-3333-3333-3333-333333333333"
	testProjectID    = "22222222-2222-2222-2222-222222222222"
	testCheckID      = "44444444-4444-4444-4444-444444444444"
	testAssignmentID = "66666666-6666-6666-6666-666666666666"
)

func TestHelloRejectsInvalidSecret(t *testing.T) {
	probes := &fakeProbeRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, fakeSecretVerifier{valid: false}, recorder)

	_, err := service.Hello(context.Background(), RuntimeStatusInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "wrong",
		},
	})

	if !errors.Is(err, domainprobe.ErrInvalidCredential) {
		t.Fatalf("expected invalid credential, got %v", err)
	}
	if probes.gotStatus.ProbeID != "" {
		t.Fatalf("expected status not to update, got %#v", probes.gotStatus)
	}
	assertRecordedProbeRuntimeEvent(t, recorder, ProbeRuntimeEvent{
		Name:      ProbeRuntimeEventHelloFailure,
		Action:    ProbeRuntimeActionHello,
		Outcome:   ProbeRuntimeOutcomeFailure,
		Reason:    ProbeRuntimeReasonInvalidCredential,
		ProbeID:   testProbeID,
		ProjectID: testProjectID,
	})
}

func TestHelloRecordsDisabledProbe(t *testing.T) {
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(
		&fakeProbeRepository{credential: domainprobe.Credential{
			ProbeID:    testProbeID,
			ProjectID:  testProjectID,
			Enabled:    false,
			SecretHash: "secret-hash",
		}},
		fakeSecretVerifier{valid: true},
		recorder,
	)

	_, err := service.Hello(context.Background(), RuntimeStatusInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
	})
	if !errors.Is(err, domainprobe.ErrProbeDisabled) {
		t.Fatalf("expected probe disabled, got %v", err)
	}
	assertRecordedProbeRuntimeEvent(t, recorder, ProbeRuntimeEvent{
		Name:      ProbeRuntimeEventHelloFailure,
		Action:    ProbeRuntimeActionHello,
		Outcome:   ProbeRuntimeOutcomeFailure,
		Reason:    ProbeRuntimeReasonProbeDisabled,
		ProbeID:   testProbeID,
		ProjectID: testProjectID,
	})
}

func TestHeartbeatUpdatesOnlineStatus(t *testing.T) {
	probes := &fakeProbeRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, fakeSecretVerifier{valid: true}, recorder)
	agentVersion := "netstamp-probe/0.1.0"
	as := "AS15169 Google LLC"

	output, err := service.Heartbeat(context.Background(), RuntimeStatusInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		AgentVersion: &agentVersion,
		AS:           &as,
	})
	if err != nil {
		t.Fatalf("expected heartbeat to succeed: %v", err)
	}
	if output.ServerTime.IsZero() {
		t.Fatal("expected server time")
	}
	if probes.gotStatus.ProbeID != testProbeID {
		t.Fatalf("expected probe id %q, got %q", testProbeID, probes.gotStatus.ProbeID)
	}
	if probes.gotStatus.State != domainprobe.StateOnline {
		t.Fatalf("expected online status, got %q", probes.gotStatus.State)
	}
	if probes.gotStatus.AgentVersion == nil || *probes.gotStatus.AgentVersion != agentVersion {
		t.Fatalf("expected agent version, got %#v", probes.gotStatus.AgentVersion)
	}
	if probes.gotStatus.AS == nil || *probes.gotStatus.AS != as {
		t.Fatalf("expected AS, got %#v", probes.gotStatus.AS)
	}
	assertNoProbeRuntimeEvents(t, recorder)
}

func TestListAssignmentsAuthenticatesProbe(t *testing.T) {
	probes := &fakeProbeRepository{
		assignments: []domainassignment.Assignment{newAssignment(testCheckID, domaincheck.TypePing)},
	}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, fakeSecretVerifier{valid: true}, recorder)

	output, err := service.ListAssignments(context.Background(), RuntimeAuthInput{
		ProbeID:    testProbeID,
		Credential: "plain-secret",
	})
	if err != nil {
		t.Fatalf("expected assignments to succeed: %v", err)
	}
	if len(output.Assignments) != 1 || output.Assignments[0].CheckID != testCheckID {
		t.Fatalf("expected assignment for check %q, got %#v", testCheckID, output.Assignments)
	}
	if output.Assignments[0].Check == nil || output.Assignments[0].Check.Type != domaincheck.TypePing {
		t.Fatalf("expected domain check data on assignment, got %#v", output.Assignments[0])
	}
	assertNoProbeRuntimeEvents(t, recorder)
}

func TestHeartbeatRecordsStatusUpdateFailure(t *testing.T) {
	recorder := &recordingProbeRuntimeEventRecorder{}
	updateErr := errors.New("update status")
	service := NewService(
		&fakeProbeRepository{updateErr: updateErr},
		fakeSecretVerifier{valid: true},
		recorder,
	)

	_, err := service.Heartbeat(context.Background(), RuntimeStatusInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
	})
	if !errors.Is(err, updateErr) {
		t.Fatalf("expected status update error, got %v", err)
	}
	assertRecordedProbeRuntimeEvent(t, recorder, ProbeRuntimeEvent{
		Name:      ProbeRuntimeEventHeartbeatFailure,
		Action:    ProbeRuntimeActionHeartbeat,
		Outcome:   ProbeRuntimeOutcomeFailure,
		Reason:    ProbeRuntimeReasonStatusUpdateFailed,
		ProbeID:   testProbeID,
		ProjectID: testProjectID,
		Err:       updateErr,
	})
}

func TestListAssignmentsRecordsAssignmentListFailure(t *testing.T) {
	recorder := &recordingProbeRuntimeEventRecorder{}
	listErr := errors.New("list assignments")
	service := NewService(
		&fakeProbeRepository{assignmentErr: listErr},
		fakeSecretVerifier{valid: true},
		recorder,
	)

	_, err := service.ListAssignments(context.Background(), RuntimeAuthInput{
		ProbeID:    testProbeID,
		Credential: "plain-secret",
	})
	if !errors.Is(err, listErr) {
		t.Fatalf("expected assignment list error, got %v", err)
	}
	assertRecordedProbeRuntimeEvent(t, recorder, ProbeRuntimeEvent{
		Name:      ProbeRuntimeEventListAssignmentsFailure,
		Action:    ProbeRuntimeActionListAssignments,
		Outcome:   ProbeRuntimeOutcomeFailure,
		Reason:    ProbeRuntimeReasonAssignmentListFailed,
		ProbeID:   testProbeID,
		ProjectID: testProjectID,
		Err:       listErr,
	})
}

func TestRuntimeRecordsProbeNotFoundFailure(t *testing.T) {
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(
		&fakeProbeRepository{credentialErr: domainprobe.ErrProbeNotFound},
		fakeSecretVerifier{valid: true},
		recorder,
	)

	_, err := service.ListAssignments(context.Background(), RuntimeAuthInput{
		ProbeID:    testProbeID,
		Credential: "plain-secret",
	})
	if !errors.Is(err, domainprobe.ErrProbeNotFound) {
		t.Fatalf("expected probe not found, got %v", err)
	}
	assertRecordedProbeRuntimeEvent(t, recorder, ProbeRuntimeEvent{
		Name:    ProbeRuntimeEventListAssignmentsFailure,
		Action:  ProbeRuntimeActionListAssignments,
		Outcome: ProbeRuntimeOutcomeFailure,
		Reason:  ProbeRuntimeReasonProbeNotFound,
		ProbeID: testProbeID,
	})
}

func TestRuntimeRecordsSecretVerifierMissing(t *testing.T) {
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(&fakeProbeRepository{}, nil, recorder)

	_, err := service.Hello(context.Background(), RuntimeStatusInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
	})
	if !errors.Is(err, errSecretVerifierMissing) {
		t.Fatalf("expected missing verifier error, got %v", err)
	}
	assertRecordedProbeRuntimeEvent(t, recorder, ProbeRuntimeEvent{
		Name:      ProbeRuntimeEventHelloFailure,
		Action:    ProbeRuntimeActionHello,
		Outcome:   ProbeRuntimeOutcomeFailure,
		Reason:    ProbeRuntimeReasonSecretVerifierMissing,
		ProbeID:   testProbeID,
		ProjectID: testProjectID,
		Err:       errSecretVerifierMissing,
	})
}

func newAssignment(checkID string, checkType domaincheck.Type) domainassignment.Assignment {
	pingConfig := domainping.DefaultConfig()

	return domainassignment.Assignment{
		ID:              testAssignmentID,
		ProjectID:       testProjectID,
		ProbeID:         testProbeID,
		CheckID:         checkID,
		CheckVersion:    "check-v1",
		SelectorVersion: "selector-v1",
		Check: &domaincheck.Check{
			ID:              checkID,
			ProjectID:       testProjectID,
			Type:            checkType,
			Target:          "1.1.1.1",
			IntervalSeconds: 30,
			PingConfig:      &pingConfig,
		},
	}
}

func assertRecordedProbeRuntimeEvent(t *testing.T, recorder *recordingProbeRuntimeEventRecorder, want ProbeRuntimeEvent) {
	t.Helper()

	if len(recorder.events) != 1 {
		t.Fatalf("expected one event, got %d: %#v", len(recorder.events), recorder.events)
	}

	got := recorder.events[0]
	if got.Name != want.Name ||
		got.Action != want.Action ||
		got.Outcome != want.Outcome ||
		got.Reason != want.Reason ||
		got.ProbeID != want.ProbeID ||
		got.ProjectID != want.ProjectID ||
		!errors.Is(got.Err, want.Err) {
		t.Fatalf("unexpected event:\n got: %#v\nwant: %#v", got, want)
	}
}

func assertNoProbeRuntimeEvents(t *testing.T, recorder *recordingProbeRuntimeEventRecorder) {
	t.Helper()

	if len(recorder.events) != 0 {
		t.Fatalf("expected no events, got %d: %#v", len(recorder.events), recorder.events)
	}
}

type recordingProbeRuntimeEventRecorder struct {
	events []ProbeRuntimeEvent
}

func (r *recordingProbeRuntimeEventRecorder) RecordProbeRuntimeEvent(_ context.Context, event ProbeRuntimeEvent) {
	r.events = append(r.events, event)
}

type fakeProbeRepository struct {
	credential    domainprobe.Credential
	credentialErr error
	gotStatus     domainprobe.Status
	updateErr     error
	assignments   []domainassignment.Assignment
	assignmentErr error
}

func (r *fakeProbeRepository) GetActiveProbeCredential(context.Context, string) (domainprobe.Credential, error) {
	if r.credentialErr != nil {
		return domainprobe.Credential{}, r.credentialErr
	}
	if r.credential.ProbeID != "" {
		return r.credential, nil
	}
	return domainprobe.Credential{
		ProbeID:    testProbeID,
		ProjectID:  testProjectID,
		Enabled:    true,
		SecretHash: "secret-hash",
	}, nil
}

func (r *fakeProbeRepository) UpdateProbeStatus(_ context.Context, input domainprobe.Status) (domainprobe.Status, error) {
	r.gotStatus = input
	if r.updateErr != nil {
		return domainprobe.Status{}, r.updateErr
	}
	return domainprobe.Status{ProbeID: input.ProbeID, State: input.State, UpdatedAt: time.Now()}, nil
}

func (r *fakeProbeRepository) ListAssignments(context.Context, string) ([]domainassignment.Assignment, error) {
	if r.assignmentErr != nil {
		return nil, r.assignmentErr
	}
	return r.assignments, nil
}

type fakeSecretVerifier struct {
	valid bool
}

func (v fakeSecretVerifier) VerifyProbeSecret(string, string) bool {
	return v.valid
}
