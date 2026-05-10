package proberuntime

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

const (
	testProbeID      = "33333333-3333-3333-3333-333333333333"
	testProjectID    = "22222222-2222-2222-2222-222222222222"
	testCheckID      = "44444444-4444-4444-4444-444444444444"
	otherCheckID     = "55555555-5555-5555-5555-555555555555"
	testAssignmentID = "66666666-6666-6666-6666-666666666666"
)

func TestHelloRejectsInvalidSecret(t *testing.T) {
	probes := &fakeProbeRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, &fakePingResultRepository{}, fakeSecretVerifier{valid: false}, recorder)

	_, err := service.Hello(context.Background(), RuntimeStatusInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "wrong",
		},
	})

	if !errors.Is(err, ErrInvalidCredential) {
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
		&fakePingResultRepository{},
		fakeSecretVerifier{valid: true},
		recorder,
	)

	_, err := service.Hello(context.Background(), RuntimeStatusInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
	})
	if !errors.Is(err, ErrProbeDisabled) {
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
	service := NewService(probes, &fakePingResultRepository{}, fakeSecretVerifier{valid: true}, recorder)
	agentVersion := "netstamp-probe/0.1.0"

	output, err := service.Heartbeat(context.Background(), RuntimeStatusInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		AgentVersion: &agentVersion,
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
	assertNoProbeRuntimeEvents(t, recorder)
}

func TestListAssignmentsAuthenticatesProbe(t *testing.T) {
	probes := &fakeProbeRepository{
		assignments: []domaincheck.Assignment{{
			ID:              "66666666-6666-6666-6666-666666666666",
			ProjectID:       testProjectID,
			ProbeID:         testProbeID,
			CheckID:         testCheckID,
			CheckVersion:    "check-v1",
			SelectorVersion: "selector-v1",
			Type:            domaincheck.TypePing,
			Target:          "1.1.1.1",
			IntervalSeconds: 30,
			PingConfig:      domainping.DefaultConfig(),
		}},
	}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, &fakePingResultRepository{}, fakeSecretVerifier{valid: true}, recorder)

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
	assertNoProbeRuntimeEvents(t, recorder)
}

func TestSubmitResultsAcceptsPingBatch(t *testing.T) {
	probes := &fakeProbeRepository{assignments: []domaincheck.Assignment{newAssignment(testCheckID, domaincheck.TypePing)}}
	results := &fakePingResultRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, results, fakeSecretVerifier{valid: true}, recorder)
	startedAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(time.Second)
	ipFamily := domainnetwork.IPFamilyInet

	output, err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		Groups: []ResultGroupInput{newResultGroupInput(testCheckID,
			newPingResultInput(startedAt, finishedAt, &ipFamily),
			newPingResultInput(startedAt.Add(30*time.Second), finishedAt.Add(30*time.Second), nil),
		)},
	})
	if err != nil {
		t.Fatalf("expected submit to succeed: %v", err)
	}
	if !output.Accepted || output.ResyncNeeded || len(output.StaleChecks) != 0 {
		t.Fatalf("expected accepted fresh output, got %#v", output)
	}
	if len(results.created) != 2 {
		t.Fatalf("expected two ping results to reach repository, got %#v", results.created)
	}
	if results.created[0].ProjectID != testProjectID || results.created[0].ProbeID != testProbeID {
		t.Fatalf("expected project/probe ids to be set from credential, got %#v", results.created[0])
	}
	assertNoProbeRuntimeEvents(t, recorder)
}

func TestSubmitResultsAcceptsStaleVersionAndRequestsResync(t *testing.T) {
	assignment := newAssignment(testCheckID, domaincheck.TypePing)
	probes := &fakeProbeRepository{assignments: []domaincheck.Assignment{assignment}}
	results := &fakePingResultRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, results, fakeSecretVerifier{valid: true}, recorder)
	startedAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)

	group := newResultGroupInput(testCheckID, newPingResultInput(startedAt, startedAt.Add(time.Second), nil))
	group.CheckVersion = "old-check-version"

	output, err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		Groups: []ResultGroupInput{group},
	})
	if err != nil {
		t.Fatalf("expected stale submit to succeed: %v", err)
	}
	if !output.Accepted || !output.ResyncNeeded {
		t.Fatalf("expected accepted resync output, got %#v", output)
	}
	if len(output.StaleChecks) != 1 || output.StaleChecks[0] != testCheckID {
		t.Fatalf("expected stale check id, got %#v", output.StaleChecks)
	}
	if len(output.Assignments) != 1 || output.Assignments[0].CheckVersion != assignment.CheckVersion {
		t.Fatalf("expected latest assignment, got %#v", output.Assignments)
	}
	if len(results.created) != 1 {
		t.Fatalf("expected stale result to be stored, got %#v", results.created)
	}
	assertNoProbeRuntimeEvents(t, recorder)
}

func TestSubmitResultsRejectsUnsupportedType(t *testing.T) {
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(&fakeProbeRepository{}, &fakePingResultRepository{}, fakeSecretVerifier{valid: true}, recorder)
	startedAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)

	_, err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		Groups: []ResultGroupInput{{
			CheckID:         testCheckID,
			Type:            domaincheck.Type("dns"),
			AssignmentID:    testAssignmentID,
			CheckVersion:    "check-v1",
			SelectorVersion: "selector-v1",
			PingResults:     []PingResultInput{newPingResultInput(startedAt, startedAt.Add(time.Second), nil)},
		}},
	})

	if !errors.Is(err, ErrUnsupportedResult) {
		t.Fatalf("expected unsupported result, got %v", err)
	}
	assertRecordedProbeRuntimeEvent(t, recorder, ProbeRuntimeEvent{
		Name:        ProbeRuntimeEventSubmitResultsFailure,
		Action:      ProbeRuntimeActionSubmitResults,
		Outcome:     ProbeRuntimeOutcomeFailure,
		Reason:      ProbeRuntimeReasonUnsupportedResult,
		ProbeID:     testProbeID,
		ProjectID:   testProjectID,
		ResultCount: intPtr(1),
	})
}

func TestSubmitResultsRejectsUnassignedCheck(t *testing.T) {
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(&fakeProbeRepository{
		assignments: []domaincheck.Assignment{newAssignment(testCheckID, domaincheck.TypePing)},
	}, &fakePingResultRepository{}, fakeSecretVerifier{valid: true}, recorder)
	startedAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)

	_, err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		Groups: []ResultGroupInput{newResultGroupInput(otherCheckID, newPingResultInput(startedAt, startedAt.Add(time.Second), nil))},
	})

	if !errors.Is(err, ErrResultConflict) {
		t.Fatalf("expected result conflict, got %v", err)
	}
	assertRecordedProbeRuntimeEvent(t, recorder, ProbeRuntimeEvent{
		Name:        ProbeRuntimeEventSubmitResultsFailure,
		Action:      ProbeRuntimeActionSubmitResults,
		Outcome:     ProbeRuntimeOutcomeFailure,
		Reason:      ProbeRuntimeReasonResultConflict,
		ProbeID:     testProbeID,
		ProjectID:   testProjectID,
		ResultCount: intPtr(1),
	})
}

func TestSubmitResultsRejectsCheckTypeMismatch(t *testing.T) {
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(&fakeProbeRepository{
		assignments: []domaincheck.Assignment{newAssignment(testCheckID, domaincheck.Type("dns"))},
	}, &fakePingResultRepository{}, fakeSecretVerifier{valid: true}, recorder)
	startedAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)

	_, err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		Groups: []ResultGroupInput{newResultGroupInput(testCheckID, newPingResultInput(startedAt, startedAt.Add(time.Second), nil))},
	})

	if !errors.Is(err, ErrResultConflict) {
		t.Fatalf("expected result conflict, got %v", err)
	}
	assertRecordedProbeRuntimeEvent(t, recorder, ProbeRuntimeEvent{
		Name:        ProbeRuntimeEventSubmitResultsFailure,
		Action:      ProbeRuntimeActionSubmitResults,
		Outcome:     ProbeRuntimeOutcomeFailure,
		Reason:      ProbeRuntimeReasonResultConflict,
		ProbeID:     testProbeID,
		ProjectID:   testProjectID,
		ResultCount: intPtr(1),
	})
}

func TestSubmitResultsRejectsInvalidPing(t *testing.T) {
	results := &fakePingResultRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(&fakeProbeRepository{
		assignments: []domaincheck.Assignment{newAssignment(testCheckID, domaincheck.TypePing)},
	}, results, fakeSecretVerifier{valid: true}, recorder)
	startedAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)

	_, err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		Groups: []ResultGroupInput{newResultGroupInput(testCheckID, newPingResultInput(startedAt, startedAt.Add(-time.Second), nil))},
	})

	if !errors.Is(err, ErrInvalidResult) {
		t.Fatalf("expected invalid result, got %v", err)
	}
	if len(results.created) != 0 {
		t.Fatalf("expected whole batch to be rejected before writes, got %#v", results.created)
	}
	assertRecordedProbeRuntimeEvent(t, recorder, ProbeRuntimeEvent{
		Name:        ProbeRuntimeEventSubmitResultsFailure,
		Action:      ProbeRuntimeActionSubmitResults,
		Outcome:     ProbeRuntimeOutcomeFailure,
		Reason:      ProbeRuntimeReasonInvalidResult,
		ProbeID:     testProbeID,
		ProjectID:   testProjectID,
		ResultCount: intPtr(1),
	})
}

func TestSubmitResultsRejectsEmptyBatch(t *testing.T) {
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(&fakeProbeRepository{}, &fakePingResultRepository{}, fakeSecretVerifier{valid: true}, recorder)

	_, err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	assertRecordedProbeRuntimeEvent(t, recorder, ProbeRuntimeEvent{
		Name:        ProbeRuntimeEventSubmitResultsFailure,
		Action:      ProbeRuntimeActionSubmitResults,
		Outcome:     ProbeRuntimeOutcomeFailure,
		Reason:      ProbeRuntimeReasonInvalidInput,
		ProbeID:     testProbeID,
		ProjectID:   testProjectID,
		ResultCount: intPtr(0),
	})
}

func TestHeartbeatRecordsStatusUpdateFailure(t *testing.T) {
	recorder := &recordingProbeRuntimeEventRecorder{}
	updateErr := errors.New("update status")
	service := NewService(
		&fakeProbeRepository{updateErr: updateErr},
		&fakePingResultRepository{},
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
		&fakePingResultRepository{},
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

func TestSubmitResultsRecordsResultWriteFailure(t *testing.T) {
	recorder := &recordingProbeRuntimeEventRecorder{}
	writeErr := errors.New("write results")
	service := NewService(
		&fakeProbeRepository{assignments: []domaincheck.Assignment{newAssignment(testCheckID, domaincheck.TypePing)}},
		&fakePingResultRepository{err: writeErr},
		fakeSecretVerifier{valid: true},
		recorder,
	)
	startedAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)

	_, err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		Groups: []ResultGroupInput{newResultGroupInput(testCheckID, newPingResultInput(startedAt, startedAt.Add(time.Second), nil))},
	})
	if !errors.Is(err, writeErr) {
		t.Fatalf("expected write error, got %v", err)
	}
	assertRecordedProbeRuntimeEvent(t, recorder, ProbeRuntimeEvent{
		Name:        ProbeRuntimeEventSubmitResultsFailure,
		Action:      ProbeRuntimeActionSubmitResults,
		Outcome:     ProbeRuntimeOutcomeFailure,
		Reason:      ProbeRuntimeReasonResultWriteFailed,
		ProbeID:     testProbeID,
		ProjectID:   testProjectID,
		ResultCount: intPtr(1),
		Err:         writeErr,
	})
}

func TestRuntimeRecordsProbeNotFoundFailure(t *testing.T) {
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(
		&fakeProbeRepository{credentialErr: ErrProbeNotFound},
		&fakePingResultRepository{},
		fakeSecretVerifier{valid: true},
		recorder,
	)

	_, err := service.ListAssignments(context.Background(), RuntimeAuthInput{
		ProbeID:    testProbeID,
		Credential: "plain-secret",
	})
	if !errors.Is(err, ErrProbeNotFound) {
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
	service := NewService(&fakeProbeRepository{}, &fakePingResultRepository{}, nil, recorder)

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

func newAssignment(checkID string, checkType domaincheck.Type) domaincheck.Assignment {
	return domaincheck.Assignment{
		ID:              testAssignmentID,
		ProjectID:       testProjectID,
		ProbeID:         testProbeID,
		CheckID:         checkID,
		CheckVersion:    "check-v1",
		SelectorVersion: "selector-v1",
		Type:            checkType,
		Target:          "1.1.1.1",
		IntervalSeconds: 30,
		PingConfig:      domainping.DefaultConfig(),
	}
}

func newResultGroupInput(checkID string, results ...PingResultInput) ResultGroupInput {
	return ResultGroupInput{
		CheckID:         checkID,
		Type:            domaincheck.TypePing,
		AssignmentID:    testAssignmentID,
		CheckVersion:    "check-v1",
		SelectorVersion: "selector-v1",
		PingResults:     results,
	}
}

func newPingResultInput(startedAt, finishedAt time.Time, ipFamily *domainnetwork.IPFamily) PingResultInput {
	return PingResultInput{
		StartedAt:     startedAt,
		FinishedAt:    finishedAt,
		DurationMs:    int32(finishedAt.Sub(startedAt).Milliseconds()),
		Status:        string(domainping.StatusSuccessful),
		SentCount:     4,
		ReceivedCount: 4,
		LossPercent:   0,
		RttMinMs:      float64Ptr(10),
		RttAvgMs:      float64Ptr(11),
		RttMedianMs:   float64Ptr(11),
		RttMaxMs:      float64Ptr(12),
		RttStddevMs:   float64Ptr(1),
		RttSamplesMs:  []float64{10, 11, 12},
		IPFamily:      ipFamily,
		Raw:           json.RawMessage(`{"runner":"test"}`),
	}
}

func float64Ptr(value float64) *float64 {
	return &value
}

func intPtr(value int) *int {
	return &value
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
		!sameIntPtr(got.ResultCount, want.ResultCount) ||
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

func sameIntPtr(left, right *int) bool {
	switch {
	case left == nil && right == nil:
		return true
	case left == nil || right == nil:
		return false
	default:
		return *left == *right
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
	gotStatus     domainprobe.UpdateStatusInput
	updateErr     error
	assignments   []domaincheck.Assignment
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

func (r *fakeProbeRepository) UpdateProbeStatus(_ context.Context, input domainprobe.UpdateStatusInput) (domainprobe.Status, error) {
	r.gotStatus = input
	if r.updateErr != nil {
		return domainprobe.Status{}, r.updateErr
	}
	return domainprobe.Status{ProbeID: input.ProbeID, State: input.State, UpdatedAt: time.Now()}, nil
}

func (r *fakeProbeRepository) ListAssignments(context.Context, string) ([]domaincheck.Assignment, error) {
	if r.assignmentErr != nil {
		return nil, r.assignmentErr
	}
	return r.assignments, nil
}

type fakePingResultRepository struct {
	created []domainping.ResultStorageInput
	err     error
}

func (r *fakePingResultRepository) CreatePingResults(_ context.Context, inputs []domainping.ResultStorageInput) error {
	if r.err != nil {
		return r.err
	}
	r.created = append(r.created, inputs...)
	return nil
}

type fakeSecretVerifier struct {
	valid bool
}

func (v fakeSecretVerifier) VerifyProbeSecret(string, string) bool {
	return v.valid
}
