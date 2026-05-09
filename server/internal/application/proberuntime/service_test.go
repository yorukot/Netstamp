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
	testProbeID   = "33333333-3333-3333-3333-333333333333"
	testProjectID = "22222222-2222-2222-2222-222222222222"
	testCheckID   = "44444444-4444-4444-4444-444444444444"
	otherCheckID  = "55555555-5555-5555-5555-555555555555"
)

func TestHelloRejectsInvalidSecret(t *testing.T) {
	probes := &fakeProbeRepository{}
	service := NewService(probes, &fakePingResultRepository{}, fakeSecretVerifier{valid: false})

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
}

func TestHeartbeatUpdatesOnlineStatus(t *testing.T) {
	probes := &fakeProbeRepository{}
	service := NewService(probes, &fakePingResultRepository{}, fakeSecretVerifier{valid: true})
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
	service := NewService(probes, &fakePingResultRepository{}, fakeSecretVerifier{valid: true})

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
}

func TestSubmitResultsAcceptsPingBatch(t *testing.T) {
	probes := &fakeProbeRepository{assignments: []domaincheck.Assignment{{
		CheckID: testCheckID,
		Type:    domaincheck.TypePing,
	}}}
	results := &fakePingResultRepository{}
	service := NewService(probes, results, fakeSecretVerifier{valid: true})
	startedAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(time.Second)
	ipFamily := domainnetwork.IPFamilyInet

	err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		Ping: []PingResultInput{
			newPingResultInput(testCheckID, startedAt, finishedAt, &ipFamily),
			newPingResultInput(testCheckID, startedAt.Add(30*time.Second), finishedAt.Add(30*time.Second), nil),
		},
	})
	if err != nil {
		t.Fatalf("expected submit to succeed: %v", err)
	}
	if len(results.created) != 2 {
		t.Fatalf("expected two ping results to reach repository, got %#v", results.created)
	}
	if results.created[0].ProjectID != testProjectID || results.created[0].ProbeID != testProbeID {
		t.Fatalf("expected project/probe ids to be set from credential, got %#v", results.created[0])
	}
}

func TestSubmitResultsRejectsUnsupportedBucket(t *testing.T) {
	service := NewService(&fakeProbeRepository{}, &fakePingResultRepository{}, fakeSecretVerifier{valid: true})

	err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		DNS: []UnsupportedResultInput{{Raw: json.RawMessage(`{}`)}},
	})

	if !errors.Is(err, ErrUnsupportedResult) {
		t.Fatalf("expected unsupported result, got %v", err)
	}
}

func TestSubmitResultsRejectsUnassignedCheck(t *testing.T) {
	service := NewService(&fakeProbeRepository{assignments: []domaincheck.Assignment{{
		CheckID: testCheckID,
		Type:    domaincheck.TypePing,
	}}}, &fakePingResultRepository{}, fakeSecretVerifier{valid: true})
	startedAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)

	err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		Ping: []PingResultInput{newPingResultInput(otherCheckID, startedAt, startedAt.Add(time.Second), nil)},
	})

	if !errors.Is(err, ErrResultConflict) {
		t.Fatalf("expected result conflict, got %v", err)
	}
}

func TestSubmitResultsRejectsCheckTypeMismatch(t *testing.T) {
	service := NewService(&fakeProbeRepository{assignments: []domaincheck.Assignment{{
		CheckID: testCheckID,
		Type:    domaincheck.Type("dns"),
	}}}, &fakePingResultRepository{}, fakeSecretVerifier{valid: true})
	startedAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)

	err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		Ping: []PingResultInput{newPingResultInput(testCheckID, startedAt, startedAt.Add(time.Second), nil)},
	})

	if !errors.Is(err, ErrResultConflict) {
		t.Fatalf("expected result conflict, got %v", err)
	}
}

func TestSubmitResultsRejectsInvalidPing(t *testing.T) {
	results := &fakePingResultRepository{}
	service := NewService(&fakeProbeRepository{assignments: []domaincheck.Assignment{{
		CheckID: testCheckID,
		Type:    domaincheck.TypePing,
	}}}, results, fakeSecretVerifier{valid: true})
	startedAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)

	err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		Ping: []PingResultInput{newPingResultInput(testCheckID, startedAt, startedAt.Add(-time.Second), nil)},
	})

	if !errors.Is(err, ErrInvalidResult) {
		t.Fatalf("expected invalid result, got %v", err)
	}
	if len(results.created) != 0 {
		t.Fatalf("expected whole batch to be rejected before writes, got %#v", results.created)
	}
}

func TestSubmitResultsRejectsEmptyBatch(t *testing.T) {
	service := NewService(&fakeProbeRepository{}, &fakePingResultRepository{}, fakeSecretVerifier{valid: true})

	err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func newPingResultInput(checkID string, startedAt, finishedAt time.Time, ipFamily *domainnetwork.IPFamily) PingResultInput {
	return PingResultInput{
		CheckID:       checkID,
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

type fakeProbeRepository struct {
	credential    domainprobe.Credential
	credentialErr error
	gotStatus     domainprobe.UpdateStatusInput
	updateErr     error
	assignments   []domaincheck.Assignment
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
	return r.assignments, nil
}

type fakePingResultRepository struct {
	created []domainping.ResultStorageInput
}

func (r *fakePingResultRepository) CreatePingResults(_ context.Context, inputs []domainping.ResultStorageInput) error {
	r.created = append(r.created, inputs...)
	return nil
}

type fakeSecretVerifier struct {
	valid bool
}

func (v fakeSecretVerifier) VerifyProbeSecret(string, string) bool {
	return v.valid
}
