package proberuntime

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humatest"

	appproberuntime "github.com/yorukot/netstamp/internal/application/proberuntime"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

const (
	handlerProbeID      = "33333333-3333-3333-3333-333333333333"
	handlerProjectID    = "22222222-2222-2222-2222-222222222222"
	handlerCheckID      = "44444444-4444-4444-4444-444444444444"
	handlerAssignmentID = "55555555-5555-5555-5555-555555555555"
)

func TestHelloRequiresProbeAuthorization(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(newRuntimeService(&handlerProbeRepository{}, &handlerPingResultRepository{}, true)).RegisterRoutes(api)

	res := api.Post("/probes/"+handlerProbeID+"/runtime/hello", map[string]any{
		"agentVersion": "netstamp-probe/0.1.0",
	})

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
}

func TestHeartbeatUpdatesRuntimeStatus(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerProbeRepository{}
	NewHandler(newRuntimeService(repo, &handlerPingResultRepository{}, true)).RegisterRoutes(api)

	res := api.Post("/probes/"+handlerProbeID+"/runtime/heartbeat", map[string]any{
		"agentVersion": "netstamp-probe/0.1.0",
		"publicV4":     "203.0.113.10",
		"publicV6":     "2001:db8::10",
		"addrs":        []string{"10.0.0.10", "fd00::10"},
	}, "Authorization: Probe plain-secret")

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	if repo.gotStatus.ProbeID != handlerProbeID {
		t.Fatalf("expected status update for probe, got %#v", repo.gotStatus)
	}
	if repo.gotStatus.PublicV4 == nil || repo.gotStatus.PublicV4.String() != "203.0.113.10" {
		t.Fatalf("expected public v4, got %#v", repo.gotStatus.PublicV4)
	}
	if len(repo.gotStatus.Addrs) != 2 {
		t.Fatalf("expected addrs to be parsed, got %#v", repo.gotStatus.Addrs)
	}
}

func TestListAssignmentsReturnsProbeAssignments(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerProbeRepository{
		assignments: []domaincheck.Assignment{{
			ID:              handlerAssignmentID,
			ProjectID:       handlerProjectID,
			ProbeID:         handlerProbeID,
			CheckID:         handlerCheckID,
			CheckVersion:    "check-v1",
			SelectorVersion: "selector-v1",
			Type:            domaincheck.TypePing,
			Target:          "1.1.1.1",
			IntervalSeconds: 30,
			PingConfig:      domainping.DefaultConfig(),
		}},
	}
	NewHandler(newRuntimeService(repo, &handlerPingResultRepository{}, true)).RegisterRoutes(api)

	res := api.Get("/probes/"+handlerProbeID+"/runtime/assignments", "Authorization: Probe plain-secret")

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	var body assignmentsOutputBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Assignments) != 1 || body.Assignments[0].CheckID != handlerCheckID {
		t.Fatalf("expected assignment response, got %#v", body.Assignments)
	}
	if body.Assignments[0].PingConfig.PacketCount != domainping.DefaultPacketCount {
		t.Fatalf("expected ping config, got %#v", body.Assignments[0].PingConfig)
	}
}

func TestSubmitResultsReturnsNoContentForPingBatch(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerProbeRepository{assignments: []domaincheck.Assignment{{
		CheckID: handlerCheckID,
		Type:    domaincheck.TypePing,
	}}}
	results := &handlerPingResultRepository{}
	NewHandler(newRuntimeService(repo, results, true)).RegisterRoutes(api)
	startedAt := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(time.Second)

	res := api.Post("/probes/"+handlerProbeID+"/runtime/results", map[string]any{
		"ping": []map[string]any{
			{
				"checkId":       handlerCheckID,
				"startedAt":     startedAt.Format(time.RFC3339Nano),
				"finishedAt":    finishedAt.Format(time.RFC3339Nano),
				"durationMs":    1000,
				"status":        "successful",
				"sentCount":     4,
				"receivedCount": 4,
				"lossPercent":   0,
				"rttMinMs":      10,
				"rttAvgMs":      11,
				"rttMedianMs":   11,
				"rttMaxMs":      12,
				"rttStddevMs":   1,
				"rttSamplesMs":  []float64{10, 11, 12},
				"resolvedIp":    "1.1.1.1",
				"ipFamily":      "inet",
				"raw":           map[string]any{"runner": "test"},
			},
			{
				"checkId":       handlerCheckID,
				"startedAt":     startedAt.Format(time.RFC3339Nano),
				"finishedAt":    finishedAt.Format(time.RFC3339Nano),
				"durationMs":    1000,
				"status":        "successful",
				"sentCount":     4,
				"receivedCount": 4,
				"lossPercent":   0,
			},
		},
	}, "Authorization: Probe plain-secret")

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", res.Code)
	}
	if len(results.created) != 2 {
		t.Fatalf("expected two repository writes, got %#v", results.created)
	}
}

func TestSubmitResultsRejectsInvalidResolvedIP(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(newRuntimeService(&handlerProbeRepository{}, &handlerPingResultRepository{}, true)).RegisterRoutes(api)
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)

	res := api.Post("/probes/"+handlerProbeID+"/runtime/results", map[string]any{
		"ping": []map[string]any{{
			"checkId":       handlerCheckID,
			"startedAt":     now.Format(time.RFC3339Nano),
			"finishedAt":    now.Add(time.Second).Format(time.RFC3339Nano),
			"durationMs":    1000,
			"status":        "successful",
			"sentCount":     4,
			"receivedCount": 4,
			"lossPercent":   0,
			"resolvedIp":    "not-an-ip",
		}},
	}, "Authorization: Probe plain-secret")

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", res.Code)
	}
}

func TestSubmitResultsRejectsUnsupportedDNSBucket(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(newRuntimeService(&handlerProbeRepository{}, &handlerPingResultRepository{}, true)).RegisterRoutes(api)

	res := api.Post("/probes/"+handlerProbeID+"/runtime/results", map[string]any{
		"dns": []map[string]any{{"checkId": handlerCheckID}},
	}, "Authorization: Probe plain-secret")

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", res.Code)
	}
}

func TestOldPingResultEndpointIsNotRegistered(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(newRuntimeService(&handlerProbeRepository{}, &handlerPingResultRepository{}, true)).RegisterRoutes(api)

	res := api.Post("/probes/"+handlerProbeID+"/runtime/results/ping", map[string]any{}, "Authorization: Probe plain-secret")

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", res.Code)
	}
}

func newRuntimeService(probes *handlerProbeRepository, results *handlerPingResultRepository, validSecret bool) *appproberuntime.Service {
	return appproberuntime.NewService(probes, results, handlerSecretVerifier{valid: validSecret}, handlerProbeRuntimeEvents{})
}

type handlerProbeRuntimeEvents struct{}

func (handlerProbeRuntimeEvents) RecordProbeRuntimeEvent(context.Context, appproberuntime.ProbeRuntimeEvent) {
}

type handlerProbeRepository struct {
	gotStatus   domainprobe.UpdateStatusInput
	assignments []domaincheck.Assignment
}

func (r *handlerProbeRepository) GetActiveProbeCredential(context.Context, string) (domainprobe.Credential, error) {
	return domainprobe.Credential{
		ProbeID:    handlerProbeID,
		ProjectID:  handlerProjectID,
		Enabled:    true,
		SecretHash: "secret-hash",
	}, nil
}

func (r *handlerProbeRepository) UpdateProbeStatus(_ context.Context, input domainprobe.UpdateStatusInput) (domainprobe.Status, error) {
	r.gotStatus = input
	return domainprobe.Status{ProbeID: input.ProbeID, State: input.State, UpdatedAt: time.Now()}, nil
}

func (r *handlerProbeRepository) ListAssignments(context.Context, string) ([]domaincheck.Assignment, error) {
	return r.assignments, nil
}

type handlerPingResultRepository struct {
	created []domainping.ResultStorageInput
}

func (r *handlerPingResultRepository) CreatePingResults(_ context.Context, inputs []domainping.ResultStorageInput) error {
	r.created = append(r.created, inputs...)
	return nil
}

type handlerSecretVerifier struct {
	valid bool
}

func (v handlerSecretVerifier) VerifyProbeSecret(string, string) bool {
	return v.valid
}
