package proberuntime

import (
	"context"
	"errors"
	"net/netip"
	"slices"
	"testing"
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

const (
	testProbeID      = "33333333-3333-3333-3333-333333333333"
	testProjectID    = "22222222-2222-2222-2222-222222222222"
	testCheckID      = "44444444-4444-4444-4444-444444444444"
	testAssignmentID = "66666666-6666-6666-6666-666666666666"
	testProbeStoreID = int64(101)
	testCheckStoreID = int64(202)
)

func TestAppendChangedAssignmentRedactsHTTPQuery(t *testing.T) {
	assignments := []domainassignment.Assignment{{
		ProjectID:      testProjectID,
		ProbeID:        testProbeID,
		CheckID:        testCheckID,
		ProbeStorageID: testProbeStoreID,
		CheckStorageID: testCheckStoreID,
		Check: &domaincheck.Check{
			ID:     testCheckID,
			Type:   domaincheck.TypeHTTP,
			Target: "https://example.com/health?token=secret",
		},
	}}

	changed := appendChangedAssignmentForStorage(nil, map[string]struct{}{}, assignments, testProbeStoreID, testCheckStoreID)
	if len(changed) != 1 || changed[0].CheckTarget != "https://example.com/health" {
		t.Fatalf("expected query-free HTTP alert target, got %#v", changed)
	}
}

func TestHelloRejectsInvalidSecret(t *testing.T) {
	probes := &fakeProbeRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, &fakePingResultRepository{}, &fakeTracerouteResultRepository{}, fakeSecretVerifier{valid: false}, recorder)

	_, err := service.Hello(context.Background(), RuntimeAuthInput{
		ProbeID:    testProbeID,
		Credential: "wrong",
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
		&fakePingResultRepository{},
		&fakeTracerouteResultRepository{},
		fakeSecretVerifier{valid: true},
		recorder,
	)

	_, err := service.Hello(context.Background(), RuntimeAuthInput{
		ProbeID:    testProbeID,
		Credential: "plain-secret",
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

func TestHelloReturnsVersionWithoutUpdatingStatus(t *testing.T) {
	probes := &fakeProbeRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, &fakePingResultRepository{}, &fakeTracerouteResultRepository{}, fakeSecretVerifier{valid: true}, recorder)

	output, err := service.Hello(context.Background(), RuntimeAuthInput{
		ProbeID:    testProbeID,
		Credential: "plain-secret",
	})
	if err != nil {
		t.Fatalf("expected hello to succeed: %v", err)
	}
	if output.ServerTime.IsZero() {
		t.Fatal("expected server time")
	}
	if output.MinimumSupportedAgentVersion != domainprobe.DefaultMinimumSupportedAgentVersion {
		t.Fatalf("expected minimum supported version %q, got %q", domainprobe.DefaultMinimumSupportedAgentVersion, output.MinimumSupportedAgentVersion)
	}
	if probes.gotStatus.ProbeID != "" {
		t.Fatalf("expected hello not to update status, got %#v", probes.gotStatus)
	}
	assertNoProbeRuntimeEvents(t, recorder)
}

func TestHeartbeatUpdatesOnlineStatus(t *testing.T) {
	probes := &fakeProbeRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, &fakePingResultRepository{}, &fakeTracerouteResultRepository{}, fakeSecretVerifier{valid: true}, recorder)
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

func TestUpdateIPFamilyCapabilitiesSetsReportedFamiliesAndClearsMissingFamilies(t *testing.T) {
	probes := &fakeProbeRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, &fakePingResultRepository{}, &fakeTracerouteResultRepository{}, fakeSecretVerifier{valid: true}, recorder)
	observedIP := netip.MustParseAddr("203.0.113.10")

	output, err := service.UpdateIPFamilyCapabilities(context.Background(), IPFamilyCapabilitiesInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		BodyPresent: true,
		ObservedIP:  &observedIP,
		Families:    []string{string(domainnetwork.IPFamilyInet)},
	})
	if err != nil {
		t.Fatalf("expected capability update to succeed: %v", err)
	}
	if output.ServerTime.IsZero() {
		t.Fatal("expected server time")
	}
	if len(probes.gotCapabilities) != 1 {
		t.Fatalf("expected one capability update, got %#v", probes.gotCapabilities)
	}
	if got := probes.gotCapabilities[0]; got.ProbeID != testProbeID || !got.UpdateV4 || !got.UpdateV6 || got.PublicV4 == nil || *got.PublicV4 != observedIP || got.PublicV6 != nil {
		t.Fatalf("unexpected capability update: %#v", got)
	}
	assertNoProbeRuntimeEvents(t, recorder)
}

func TestUpdateIPFamilyCapabilitiesInfersSupportedFamilyFromObservedIP(t *testing.T) {
	probes := &fakeProbeRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, &fakePingResultRepository{}, &fakeTracerouteResultRepository{}, fakeSecretVerifier{valid: true}, recorder)
	observedIP := netip.MustParseAddr("2001:db8::10")

	_, err := service.UpdateIPFamilyCapabilities(context.Background(), IPFamilyCapabilitiesInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		ObservedIP: &observedIP,
	})
	if err != nil {
		t.Fatalf("expected capability inference to succeed: %v", err)
	}
	if len(probes.gotCapabilities) != 1 {
		t.Fatalf("expected one capability update, got %#v", probes.gotCapabilities)
	}
	if got := probes.gotCapabilities[0]; got.ProbeID != testProbeID || got.UpdateV4 || !got.UpdateV6 || got.PublicV4 != nil || got.PublicV6 == nil || *got.PublicV6 != observedIP {
		t.Fatalf("unexpected inferred capability: %#v", got)
	}
	assertNoProbeRuntimeEvents(t, recorder)
}

func TestUpdateIPFamilyCapabilitiesNoopsWhenNoBodyAndNoObservedIP(t *testing.T) {
	probes := &fakeProbeRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, &fakePingResultRepository{}, &fakeTracerouteResultRepository{}, fakeSecretVerifier{valid: true}, recorder)

	_, err := service.UpdateIPFamilyCapabilities(context.Background(), IPFamilyCapabilitiesInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
	})
	if err != nil {
		t.Fatalf("expected empty capability inference to succeed: %v", err)
	}
	if len(probes.gotCapabilities) != 0 {
		t.Fatalf("expected no capability updates, got %#v", probes.gotCapabilities)
	}
	assertNoProbeRuntimeEvents(t, recorder)
}

func TestListAssignmentsAuthenticatesProbe(t *testing.T) {
	probes := &fakeProbeRepository{
		assignments: []domainassignment.Assignment{newAssignment(testCheckID, domaincheck.TypePing)},
	}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, &fakePingResultRepository{}, &fakeTracerouteResultRepository{}, fakeSecretVerifier{valid: true}, recorder)

	output, err := service.ListAssignments(context.Background(), RuntimeAuthInput{
		ProbeID:    testProbeID,
		Credential: "plain-secret",
	})
	if err != nil {
		t.Fatalf("expected assignments to succeed: %v", err)
	}
	if len(output.Assignments) != 1 || output.Assignments[0].Check == nil || output.Assignments[0].Check.ID != testCheckID {
		t.Fatalf("expected assignment for check %q, got %#v", testCheckID, output.Assignments)
	}
	if output.ServerTime.IsZero() {
		t.Fatal("expected server time")
	}
	if output.Assignments[0].Check == nil || output.Assignments[0].Check.Type != domaincheck.TypePing {
		t.Fatalf("expected domain check data on assignment, got %#v", output.Assignments[0])
	}
	assertNoProbeRuntimeEvents(t, recorder)
}

func TestSubmitResultsWritesAssignedPingResults(t *testing.T) {
	assignment := newAssignment(testCheckID, domaincheck.TypePing)
	probes := &fakeProbeRepository{assignments: []domainassignment.Assignment{assignment}}
	pings := &fakePingResultRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, pings, &fakeTracerouteResultRepository{}, fakeSecretVerifier{valid: true}, recorder)
	startedAt := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(time.Second)

	output, err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		Results: []RuntimeResultGroupInput{{
			CheckID: testCheckID,
			Type:    string(domaincheck.TypePing),
			Ping: []PingResultInput{{
				StartedAt:     startedAt,
				FinishedAt:    finishedAt,
				DurationMs:    1000,
				Status:        string(domainping.StatusSuccessful),
				SentCount:     4,
				ReceivedCount: 4,
				LossPercent:   0,
			}},
		}},
	})
	if err != nil {
		t.Fatalf("expected submit results to succeed: %v", err)
	}
	if output.Accepted != 1 || output.ServerTime.IsZero() {
		t.Fatalf("unexpected output: %#v", output)
	}
	if len(pings.gotInputs) != 1 {
		t.Fatalf("expected one ping result, got %#v", pings.gotInputs)
	}
	got := pings.gotInputs[0]
	if got.ProbeStorageID != testProbeStoreID || got.CheckStorageID != testCheckStoreID {
		t.Fatalf("expected storage identity on result, got %#v", got)
	}
	if !got.StartedAt.Equal(startedAt) || got.Status != domainping.StatusSuccessful {
		t.Fatalf("unexpected stored ping result: %#v", got)
	}
	assertNoProbeRuntimeEvents(t, recorder)
}

func TestSubmitResultsWritesAssignedTracerouteResults(t *testing.T) {
	assignment := newAssignment(testCheckID, domaincheck.TypeTraceroute)
	probes := &fakeProbeRepository{assignments: []domainassignment.Assignment{assignment}}
	pings := &fakePingResultRepository{}
	traceroutes := &fakeTracerouteResultRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(probes, pings, traceroutes, fakeSecretVerifier{valid: true}, recorder)
	startedAt := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(4 * time.Second)
	addr := netip.MustParseAddr("192.0.2.1")
	resolved := netip.MustParseAddr("93.184.216.34")
	ipFamily := "inet"
	hostname := "gateway.local"
	rttMin := 1.5
	rttAvg := 1.7
	rttMedian := 1.7
	rttMax := 1.9
	rttStddev := 0.2

	output, err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		Results: []RuntimeResultGroupInput{{
			CheckID: testCheckID,
			Type:    string(domaincheck.TypeTraceroute),
			Traceroute: []TracerouteResultInput{{
				StartedAt:          startedAt,
				FinishedAt:         finishedAt,
				DurationMs:         4000,
				Status:             string(domaintraceroute.StatusPartial),
				ResolvedIP:         &resolved,
				IPFamily:           &ipFamily,
				DestinationReached: false,
				HopCount:           1,
				Hops: []TracerouteHopInput{{
					HopIndex:      1,
					Address:       &addr,
					Hostname:      &hostname,
					SentCount:     3,
					ReceivedCount: 3,
					LossPercent:   0,
					RttMinMs:      &rttMin,
					RttAvgMs:      &rttAvg,
					RttMedianMs:   &rttMedian,
					RttMaxMs:      &rttMax,
					RttStddevMs:   &rttStddev,
					RttSamplesMs:  []float64{1.5, 1.7, 1.9},
				}},
			}},
		}},
	})
	if err != nil {
		t.Fatalf("expected submit results to succeed: %v", err)
	}
	if output.Accepted != 1 || output.ServerTime.IsZero() {
		t.Fatalf("unexpected output: %#v", output)
	}
	if len(pings.gotInputs) != 0 {
		t.Fatalf("expected no ping writes, got %#v", pings.gotInputs)
	}
	if len(traceroutes.gotInputs) != 1 {
		t.Fatalf("expected one traceroute result, got %#v", traceroutes.gotInputs)
	}
	got := traceroutes.gotInputs[0]
	if got.ProbeStorageID != testProbeStoreID || got.CheckStorageID != testCheckStoreID {
		t.Fatalf("expected storage identity on result, got %#v", got)
	}
	if !got.StartedAt.Equal(startedAt) || got.Status != domaintraceroute.StatusPartial {
		t.Fatalf("unexpected stored traceroute result: %#v", got)
	}
	if len(got.Hops) != 1 {
		t.Fatalf("expected one traceroute hop, got %#v", got.Hops)
	}
	hop := got.Hops[0]
	if hop.HopIndex != 1 || hop.Address == nil || *hop.Address != addr || !slices.Equal(hop.RttSamplesMs, []float64{1.5, 1.7, 1.9}) {
		t.Fatalf("unexpected stored traceroute hop: %#v", hop)
	}
	assertNoProbeRuntimeEvents(t, recorder)
}

func TestSubmitResultsWritesAssignedTCPResults(t *testing.T) {
	assignment := newAssignment(testCheckID, domaincheck.TypeTCP)
	probes := &fakeProbeRepository{assignments: []domainassignment.Assignment{assignment}}
	tcps := &fakeTCPResultRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewServiceWithTCP(probes, &fakePingResultRepository{}, tcps, &fakeTracerouteResultRepository{}, fakeSecretVerifier{valid: true}, recorder)
	startedAt := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(42 * time.Millisecond)
	resolved := netip.MustParseAddr("93.184.216.34")
	ipFamily := "inet"
	connectDurationMs := 42.0

	output, err := service.SubmitResults(context.Background(), SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		Results: []RuntimeResultGroupInput{{
			CheckID: testCheckID,
			Type:    string(domaincheck.TypeTCP),
			TCP: []TCPResultInput{{
				StartedAt:         startedAt,
				FinishedAt:        finishedAt,
				DurationMs:        42,
				Status:            string(domaintcp.StatusSuccessful),
				ConnectDurationMs: &connectDurationMs,
				ResolvedIP:        &resolved,
				IPFamily:          &ipFamily,
			}},
		}},
	})
	if err != nil {
		t.Fatalf("expected submit results to succeed: %v", err)
	}
	if output.Accepted != 1 || output.ServerTime.IsZero() {
		t.Fatalf("unexpected output: %#v", output)
	}
	if len(tcps.gotInputs) != 1 {
		t.Fatalf("expected one tcp result, got %#v", tcps.gotInputs)
	}
	got := tcps.gotInputs[0]
	if got.ProbeStorageID != testProbeStoreID || got.CheckStorageID != testCheckStoreID {
		t.Fatalf("expected storage identity on result, got %#v", got)
	}
	if !got.StartedAt.Equal(startedAt) || got.Status != domaintcp.StatusSuccessful {
		t.Fatalf("unexpected stored tcp result: %#v", got)
	}
	if got.ConnectDurationMs == nil || *got.ConnectDurationMs != connectDurationMs {
		t.Fatalf("expected connect duration, got %#v", got.ConnectDurationMs)
	}
	assertNoProbeRuntimeEvents(t, recorder)
}

func TestSubmitResultsRejectsUnassignedCheck(t *testing.T) {
	pings := &fakePingResultRepository{}
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(&fakeProbeRepository{}, pings, &fakeTracerouteResultRepository{}, fakeSecretVerifier{valid: true}, recorder)

	_, err := service.SubmitResults(context.Background(), validSubmitResultsInput())
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if len(pings.gotInputs) != 0 {
		t.Fatalf("expected no ping writes, got %#v", pings.gotInputs)
	}
	assertRecordedProbeRuntimeEvent(t, recorder, ProbeRuntimeEvent{
		Name:      ProbeRuntimeEventSubmitResultsFailure,
		Action:    ProbeRuntimeActionSubmitResults,
		Outcome:   ProbeRuntimeOutcomeFailure,
		Reason:    ProbeRuntimeReasonInvalidInput,
		ProbeID:   testProbeID,
		ProjectID: testProjectID,
	})
}

func TestSubmitResultsRejectsTypeMismatch(t *testing.T) {
	assignment := newAssignment(testCheckID, domaincheck.TypePing)
	assignment.Check.Type = domaincheck.Type("http")
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(
		&fakeProbeRepository{assignments: []domainassignment.Assignment{assignment}},
		&fakePingResultRepository{},
		&fakeTracerouteResultRepository{},
		fakeSecretVerifier{valid: true},
		recorder,
	)

	_, err := service.SubmitResults(context.Background(), validSubmitResultsInput())
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	assertRecordedProbeRuntimeEvent(t, recorder, ProbeRuntimeEvent{
		Name:      ProbeRuntimeEventSubmitResultsFailure,
		Action:    ProbeRuntimeActionSubmitResults,
		Outcome:   ProbeRuntimeOutcomeFailure,
		Reason:    ProbeRuntimeReasonInvalidInput,
		ProbeID:   testProbeID,
		ProjectID: testProjectID,
	})
}

func TestSubmitResultsRecordsWriteFailure(t *testing.T) {
	writeErr := errors.New("write results")
	recorder := &recordingProbeRuntimeEventRecorder{}
	service := NewService(
		&fakeProbeRepository{assignments: []domainassignment.Assignment{newAssignment(testCheckID, domaincheck.TypePing)}},
		&fakePingResultRepository{err: writeErr},
		&fakeTracerouteResultRepository{},
		fakeSecretVerifier{valid: true},
		recorder,
	)

	_, err := service.SubmitResults(context.Background(), validSubmitResultsInput())
	if !errors.Is(err, writeErr) {
		t.Fatalf("expected write error, got %v", err)
	}
	assertRecordedProbeRuntimeEvent(t, recorder, ProbeRuntimeEvent{
		Name:      ProbeRuntimeEventSubmitResultsFailure,
		Action:    ProbeRuntimeActionSubmitResults,
		Outcome:   ProbeRuntimeOutcomeFailure,
		Reason:    ProbeRuntimeReasonResultWriteFailed,
		ProbeID:   testProbeID,
		ProjectID: testProjectID,
		Err:       writeErr,
	})
}

func TestHeartbeatRecordsStatusUpdateFailure(t *testing.T) {
	recorder := &recordingProbeRuntimeEventRecorder{}
	updateErr := errors.New("update status")
	service := NewService(
		&fakeProbeRepository{updateErr: updateErr},
		&fakePingResultRepository{},
		&fakeTracerouteResultRepository{},
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
		&fakeTracerouteResultRepository{},
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
		&fakePingResultRepository{},
		&fakeTracerouteResultRepository{},
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
	service := NewService(&fakeProbeRepository{}, &fakePingResultRepository{}, &fakeTracerouteResultRepository{}, nil, recorder)

	_, err := service.Hello(context.Background(), RuntimeAuthInput{
		ProbeID:    testProbeID,
		Credential: "plain-secret",
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
	tcpConfig := domaintcp.DefaultConfig()
	tracerouteConfig := domaintraceroute.DefaultConfig()
	var pingConfigPtr *domainping.Config
	var tcpConfigPtr *domaintcp.Config
	var tracerouteConfigPtr *domaintraceroute.Config
	switch checkType {
	case domaincheck.TypePing:
		pingConfigPtr = &pingConfig
	case domaincheck.TypeTCP:
		tcpConfigPtr = &tcpConfig
	case domaincheck.TypeTraceroute:
		tracerouteConfigPtr = &tracerouteConfig
	}

	return domainassignment.Assignment{
		ID:              testAssignmentID,
		ProjectID:       testProjectID,
		ProbeStorageID:  testProbeStoreID,
		CheckStorageID:  testCheckStoreID,
		CheckVersion:    "check-v1",
		SelectorVersion: "selector-v1",
		Check: &domaincheck.Check{
			ID:               checkID,
			ProjectID:        testProjectID,
			Type:             checkType,
			Target:           "1.1.1.1",
			IntervalSeconds:  30,
			PingConfig:       pingConfigPtr,
			TCPConfig:        tcpConfigPtr,
			TracerouteConfig: tracerouteConfigPtr,
		},
	}
}

func validSubmitResultsInput() SubmitResultsInput {
	startedAt := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)

	return SubmitResultsInput{
		RuntimeAuthInput: RuntimeAuthInput{
			ProbeID:    testProbeID,
			Credential: "plain-secret",
		},
		Results: []RuntimeResultGroupInput{{
			CheckID: testCheckID,
			Type:    string(domaincheck.TypePing),
			Ping: []PingResultInput{{
				StartedAt:     startedAt,
				FinishedAt:    startedAt.Add(time.Second),
				DurationMs:    1000,
				Status:        string(domainping.StatusSuccessful),
				SentCount:     4,
				ReceivedCount: 4,
				LossPercent:   0,
			}},
		}},
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
	credential      domainprobe.Credential
	credentialErr   error
	gotStatus       domainprobe.Status
	gotCapabilities []domainprobe.IPFamilyCapabilities
	updateErr       error
	assignments     []domainassignment.Assignment
	assignmentErr   error
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

func (r *fakeProbeRepository) UpdateProbeIPFamilyCapabilities(_ context.Context, input domainprobe.IPFamilyCapabilities) (domainprobe.Status, error) {
	r.gotCapabilities = append(r.gotCapabilities, input)
	if r.updateErr != nil {
		return domainprobe.Status{}, r.updateErr
	}
	return domainprobe.Status{ProbeID: input.ProbeID, UpdatedAt: time.Now()}, nil
}

func (r *fakeProbeRepository) ListAssignments(context.Context, string) ([]domainassignment.Assignment, error) {
	if r.assignmentErr != nil {
		return nil, r.assignmentErr
	}
	return r.assignments, nil
}

func (r *fakeProbeRepository) ListActiveAssignmentsForProbeChecks(_ context.Context, _ string, checkIDs []string) ([]domainassignment.Assignment, error) {
	if r.assignmentErr != nil {
		return nil, r.assignmentErr
	}

	allowed := make(map[string]struct{}, len(checkIDs))
	for _, checkID := range checkIDs {
		allowed[checkID] = struct{}{}
	}

	assignments := make([]domainassignment.Assignment, 0, len(r.assignments))
	for _, assignment := range r.assignments {
		if assignment.Check != nil {
			if _, ok := allowed[assignment.Check.ID]; ok {
				assignments = append(assignments, assignment)
			}
		}
	}

	return assignments, nil
}

type fakePingResultRepository struct {
	gotInputs []domainping.ResultStorageInput
	err       error
}

func (r *fakePingResultRepository) CreatePingResults(_ context.Context, inputs []domainping.ResultStorageInput) ([]domainping.ResultStorageInput, error) {
	r.gotInputs = append([]domainping.ResultStorageInput(nil), inputs...)
	if r.err != nil {
		return nil, r.err
	}
	return inputs, nil
}

type fakeTCPResultRepository struct {
	gotInputs []domaintcp.ResultStorageInput
	err       error
}

func (r *fakeTCPResultRepository) CreateTCPResults(_ context.Context, inputs []domaintcp.ResultStorageInput) ([]domaintcp.ResultStorageInput, error) {
	r.gotInputs = append([]domaintcp.ResultStorageInput(nil), inputs...)
	if r.err != nil {
		return nil, r.err
	}
	return inputs, nil
}

type fakeTracerouteResultRepository struct {
	gotInputs []domaintraceroute.ResultStorageInput
	err       error
}

func (r *fakeTracerouteResultRepository) CreateTracerouteResults(_ context.Context, inputs []domaintraceroute.ResultStorageInput) error {
	r.gotInputs = append([]domaintraceroute.ResultStorageInput(nil), inputs...)
	return r.err
}

type fakeSecretVerifier struct {
	valid bool
}

func (v fakeSecretVerifier) VerifyProbeSecret(string, string) bool {
	return v.valid
}
