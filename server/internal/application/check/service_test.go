package check

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

const (
	testProjectID = "22222222-2222-2222-2222-222222222222"
	testCheckID   = "33333333-3333-3333-3333-333333333333"
	testLabelID   = "44444444-4444-4444-4444-444444444444"
)

func TestListChecksAllowsProjectMember(t *testing.T) {
	repo := &fakeCheckRepository{
		checks: []domaincheck.Check{newFakeCheck(testProjectID, testCheckID)},
	}
	projectAccess := &fakeProjectAccess{role: domainproject.RoleViewer}
	recorder := &recordingCheckEventRecorder{}
	service := NewService(repo, projectAccess, &fakeLabelAccess{}, recorder)

	checks, err := service.ListChecks(context.Background(), ListChecksInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
	})
	if err != nil {
		t.Fatalf("list checks: %v", err)
	}
	if len(checks) != 1 || checks[0].ID != testCheckID {
		t.Fatalf("expected checks, got %#v", checks)
	}
	if repo.gotListProjectID != testProjectID {
		t.Fatalf("expected list project id, got %q", repo.gotListProjectID)
	}
	if projectAccess.gotRoleProjectID != "" {
		t.Fatalf("expected list not to require role lookup, got %q", projectAccess.gotRoleProjectID)
	}
	assertNoCheckEvents(t, recorder)
}

func TestCreateCheckNormalizesInputAndDefaultsPingConfig(t *testing.T) {
	repo := &fakeCheckRepository{}
	labelAccess := &fakeLabelAccess{
		labels: []domainlabel.Label{newFakeLabel(testProjectID, testLabelID)},
	}
	recorder := &recordingCheckEventRecorder{}
	service := NewService(repo, &fakeProjectAccess{role: domainproject.RoleEditor}, labelAccess, recorder)

	description := " public API "
	check, err := service.CreateCheck(context.Background(), CreateCheckInput{
		CurrentUserID:   "user-1",
		ProjectRef:      "engineering",
		Name:            " api-latency ",
		Type:            "ping",
		Target:          " api.netstamp.io ",
		Selector:        map[string]any{"label": "edge"},
		Description:     &description,
		IntervalSeconds: 30,
		LabelIDs:        []string{testLabelID, testLabelID},
	})
	if err != nil {
		t.Fatalf("create check: %v", err)
	}
	if check.Name != "api-latency" || check.Target != "api.netstamp.io" {
		t.Fatalf("expected normalized check, got %#v", check)
	}
	if check.Description == nil || *check.Description != "public API" {
		t.Fatalf("expected trimmed description, got %#v", check.Description)
	}
	if check.PingConfig.PacketCount != defaultPacketCount ||
		check.PingConfig.PacketSizeBytes != defaultPacketSizeBytes ||
		check.PingConfig.TimeoutMs != defaultTimeoutMs {
		t.Fatalf("expected default ping config, got %#v", check.PingConfig)
	}
	if len(check.Labels) != 1 || check.Labels[0].ID != testLabelID {
		t.Fatalf("expected resolved label on output, got %#v", check.Labels)
	}
	if len(labelAccess.gotLabelIDs) != 1 || labelAccess.gotLabelIDs[0] != testLabelID {
		t.Fatalf("expected deduplicated label IDs, got %#v", labelAccess.gotLabelIDs)
	}
	if repo.gotCreate.Name != "api-latency" || repo.gotCreate.Target != "api.netstamp.io" {
		t.Fatalf("expected normalized create input, got %#v", repo.gotCreate)
	}
	var selector map[string]any
	if err := json.Unmarshal(repo.gotCreate.Selector, &selector); err != nil {
		t.Fatalf("decode selector: %v", err)
	}
	if selector["label"] != "edge" {
		t.Fatalf("expected selector to be persisted, got %#v", selector)
	}
	assertRecordedCheckEvent(t, recorder, CheckEvent{
		Name:        CheckEventCreateSuccess,
		Action:      CheckActionCreate,
		Outcome:     CheckOutcomeSuccess,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		ProjectSlug: "engineering",
		CheckID:     testCheckID,
	})
}

func TestCreateCheckRejectsViewerBeforeLabelLookup(t *testing.T) {
	repo := &fakeCheckRepository{}
	labelAccess := &fakeLabelAccess{}
	recorder := &recordingCheckEventRecorder{}
	service := NewService(repo, &fakeProjectAccess{role: domainproject.RoleViewer}, labelAccess, recorder)

	_, err := service.CreateCheck(context.Background(), CreateCheckInput{
		CurrentUserID:   "user-1",
		ProjectRef:      "engineering",
		Name:            "api-latency",
		Type:            "ping",
		Target:          "api.netstamp.io",
		IntervalSeconds: 30,
		LabelIDs:        []string{testLabelID},
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if labelAccess.gotProjectID != "" {
		t.Fatalf("expected label access not to be called, got project id %q", labelAccess.gotProjectID)
	}
	if repo.gotCreate.Name != "" {
		t.Fatalf("expected create not to be called, got %#v", repo.gotCreate)
	}
	assertRecordedCheckEvent(t, recorder, CheckEvent{
		Name:        CheckEventCreateFailure,
		Action:      CheckActionCreate,
		Outcome:     CheckOutcomeFailure,
		Reason:      CheckReasonForbidden,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		ProjectSlug: "engineering",
	})
}

func TestUpdateCheckPreservesExistingFieldsAndReplacesLabels(t *testing.T) {
	repo := &fakeCheckRepository{
		check: newFakeCheck(testProjectID, testCheckID),
	}
	labelIDs := []string{}
	intervalSeconds := 60
	timeoutMs := 1500
	recorder := &recordingCheckEventRecorder{}
	service := NewService(repo, &fakeProjectAccess{role: domainproject.RoleAdmin}, &fakeLabelAccess{}, recorder)

	check, err := service.UpdateCheck(context.Background(), UpdateCheckInput{
		CurrentUserID:   "user-1",
		ProjectRef:      "engineering",
		CheckID:         testCheckID,
		IntervalSeconds: &intervalSeconds,
		TimeoutMs:       &timeoutMs,
		LabelIDs:        &labelIDs,
	})
	if err != nil {
		t.Fatalf("update check: %v", err)
	}
	if repo.gotUpdate.Name != "api-latency" || repo.gotUpdate.Target != "api.netstamp.io" {
		t.Fatalf("expected existing fields to be preserved, got %#v", repo.gotUpdate)
	}
	if repo.gotUpdate.IntervalSeconds != 60 || repo.gotUpdate.PingConfig.TimeoutMs != 1500 {
		t.Fatalf("expected updated schedule/config, got %#v", repo.gotUpdate)
	}
	if !repo.gotUpdate.ReplaceLabels || len(repo.gotUpdate.LabelIDs) != 0 {
		t.Fatalf("expected labels to be cleared, got replace=%t ids=%#v", repo.gotUpdate.ReplaceLabels, repo.gotUpdate.LabelIDs)
	}
	if len(check.Labels) != 0 {
		t.Fatalf("expected labels to be cleared on output, got %#v", check.Labels)
	}
	assertRecordedCheckEvent(t, recorder, CheckEvent{
		Name:        CheckEventUpdateSuccess,
		Action:      CheckActionUpdate,
		Outcome:     CheckOutcomeSuccess,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		ProjectSlug: "engineering",
		CheckID:     testCheckID,
	})
}

func TestUpdateCheckRejectsEmptyPatch(t *testing.T) {
	repo := &fakeCheckRepository{}
	recorder := &recordingCheckEventRecorder{}
	service := NewService(repo, &fakeProjectAccess{role: domainproject.RoleOwner}, &fakeLabelAccess{}, recorder)

	_, err := service.UpdateCheck(context.Background(), UpdateCheckInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		CheckID:       testCheckID,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if repo.gotGetCheckID != "" || repo.gotUpdate.CheckID != "" {
		t.Fatalf("expected repository not to be called, got get=%q update=%#v", repo.gotGetCheckID, repo.gotUpdate)
	}
	assertRecordedCheckEvent(t, recorder, CheckEvent{
		Name:        CheckEventUpdateFailure,
		Action:      CheckActionUpdate,
		Outcome:     CheckOutcomeFailure,
		Reason:      CheckReasonInvalidInput,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		ProjectSlug: "engineering",
		CheckID:     testCheckID,
	})
}

func TestDeleteCheckRequiresManager(t *testing.T) {
	repo := &fakeCheckRepository{}
	recorder := &recordingCheckEventRecorder{}
	service := NewService(repo, &fakeProjectAccess{role: domainproject.RoleAdmin}, &fakeLabelAccess{}, recorder)

	err := service.DeleteCheck(context.Background(), GetCheckInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		CheckID:       testCheckID,
	})
	if err != nil {
		t.Fatalf("delete check: %v", err)
	}
	if repo.gotDeleteProjectID != testProjectID || repo.gotDeleteCheckID != testCheckID {
		t.Fatalf("expected delete ids, got project=%q check=%q", repo.gotDeleteProjectID, repo.gotDeleteCheckID)
	}
	assertRecordedCheckEvent(t, recorder, CheckEvent{
		Name:        CheckEventDeleteSuccess,
		Action:      CheckActionDelete,
		Outcome:     CheckOutcomeSuccess,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		ProjectSlug: "engineering",
		CheckID:     testCheckID,
	})
}

func assertRecordedCheckEvent(t *testing.T, recorder *recordingCheckEventRecorder, want CheckEvent) {
	t.Helper()

	if len(recorder.events) != 1 {
		t.Fatalf("expected one event, got %d: %#v", len(recorder.events), recorder.events)
	}

	got := recorder.events[0]
	if got.Name != want.Name ||
		got.Action != want.Action ||
		got.Outcome != want.Outcome ||
		got.Reason != want.Reason ||
		got.ActorUserID != want.ActorUserID ||
		got.ProjectID != want.ProjectID ||
		got.ProjectRef != want.ProjectRef ||
		got.ProjectSlug != want.ProjectSlug ||
		got.CheckID != want.CheckID ||
		!errors.Is(got.Err, want.Err) {
		t.Fatalf("unexpected event:\n got: %#v\nwant: %#v", got, want)
	}
}

func assertNoCheckEvents(t *testing.T, recorder *recordingCheckEventRecorder) {
	t.Helper()

	if len(recorder.events) != 0 {
		t.Fatalf("expected no events, got %d: %#v", len(recorder.events), recorder.events)
	}
}

type recordingCheckEventRecorder struct {
	events []CheckEvent
}

func (r *recordingCheckEventRecorder) RecordCheckEvent(_ context.Context, event CheckEvent) {
	r.events = append(r.events, event)
}

type fakeCheckRepository struct {
	checks             []domaincheck.Check
	check              domaincheck.Check
	gotListProjectID   string
	listErr            error
	gotGetProjectID    string
	gotGetCheckID      string
	getErr             error
	gotCreate          domaincheck.CreateCheckStorageInput
	createErr          error
	gotUpdate          domaincheck.UpdateCheckStorageInput
	updateErr          error
	gotDeleteProjectID string
	gotDeleteCheckID   string
	deleteErr          error
}

func (r *fakeCheckRepository) ListChecks(_ context.Context, projectID string) ([]domaincheck.Check, error) {
	r.gotListProjectID = projectID
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.checks, nil
}

func (r *fakeCheckRepository) GetCheck(_ context.Context, projectID, checkID string) (domaincheck.Check, error) {
	r.gotGetProjectID = projectID
	r.gotGetCheckID = checkID
	if r.getErr != nil {
		return domaincheck.Check{}, r.getErr
	}
	if r.check.ID != "" {
		return r.check, nil
	}
	return newFakeCheck(projectID, checkID), nil
}

func (r *fakeCheckRepository) CreateCheck(_ context.Context, input domaincheck.CreateCheckStorageInput) (domaincheck.Check, error) {
	r.gotCreate = input
	if r.createErr != nil {
		return domaincheck.Check{}, r.createErr
	}
	check := newFakeCheck(input.ProjectID, testCheckID)
	check.Name = input.Name
	check.Type = input.Type
	check.Target = input.Target
	check.Selector = input.Selector
	check.Description = input.Description
	check.IntervalSeconds = input.IntervalSeconds
	check.PingConfig = input.PingConfig
	return check, nil
}

func (r *fakeCheckRepository) UpdateCheck(_ context.Context, input domaincheck.UpdateCheckStorageInput) (domaincheck.Check, error) {
	r.gotUpdate = input
	if r.updateErr != nil {
		return domaincheck.Check{}, r.updateErr
	}
	check := newFakeCheck(input.ProjectID, input.CheckID)
	check.Name = input.Name
	check.Type = input.Type
	check.Target = input.Target
	check.Selector = input.Selector
	check.Description = input.Description
	check.IntervalSeconds = input.IntervalSeconds
	check.PingConfig = input.PingConfig
	if input.ReplaceLabels {
		check.Labels = nil
	}
	return check, nil
}

func (r *fakeCheckRepository) SoftDeleteCheck(_ context.Context, projectID, checkID string) error {
	r.gotDeleteProjectID = projectID
	r.gotDeleteCheckID = checkID
	return r.deleteErr
}

type fakeProjectAccess struct {
	role             domainproject.Role
	projectErr       error
	roleErr          error
	gotProjectRef    string
	gotUserID        string
	gotRoleProjectID string
}

func (r *fakeProjectAccess) GetProjectForUser(_ context.Context, projectRef, userID string) (domainproject.Project, error) {
	r.gotProjectRef = projectRef
	r.gotUserID = userID
	if r.projectErr != nil {
		return domainproject.Project{}, r.projectErr
	}
	return domainproject.Project{ID: testProjectID, Slug: "engineering"}, nil
}

func (r *fakeProjectAccess) GetMemberRole(_ context.Context, projectID, userID string) (domainproject.Role, error) {
	r.gotRoleProjectID = projectID
	r.gotUserID = userID
	if r.roleErr != nil {
		return "", r.roleErr
	}
	if r.role != "" {
		return r.role, nil
	}
	return domainproject.RoleOwner, nil
}

type fakeLabelAccess struct {
	labels       []domainlabel.Label
	err          error
	gotProjectID string
	gotLabelIDs  []string
}

func (r *fakeLabelAccess) GetActiveLabelsByIDsForProject(_ context.Context, projectID string, labelIDs []string) ([]domainlabel.Label, error) {
	r.gotProjectID = projectID
	r.gotLabelIDs = append([]string(nil), labelIDs...)
	if r.err != nil {
		return nil, r.err
	}
	return r.labels, nil
}

func newFakeCheck(projectID, checkID string) domaincheck.Check {
	return domaincheck.Check{
		ID:              checkID,
		ProjectID:       projectID,
		Name:            "api-latency",
		Type:            domaincheck.TypePing,
		Target:          "api.netstamp.io",
		Selector:        json.RawMessage(`{}`),
		IntervalSeconds: 30,
		PingConfig: domaincheck.PingConfig{
			PacketCount:     4,
			PacketSizeBytes: 56,
			TimeoutMs:       3000,
		},
		Labels:    []domainlabel.Label{newFakeLabel(projectID, testLabelID)},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func newFakeLabel(projectID, labelID string) domainlabel.Label {
	return domainlabel.Label{
		ID:        labelID,
		ProjectID: projectID,
		Key:       "region",
		Value:     "tokyo",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
