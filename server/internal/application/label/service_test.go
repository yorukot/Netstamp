package label

import (
	"context"
	"errors"
	"testing"
	"time"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

const (
	testProjectID = "22222222-2222-2222-2222-222222222222"
	testLabelID   = "33333333-3333-3333-3333-333333333333"
)

func TestListLabelsAllowsProjectMember(t *testing.T) {
	repo := &fakeLabelRepository{
		labels: []domainlabel.Label{{ID: testLabelID, ProjectID: testProjectID, Key: "region", Value: "tokyo"}},
	}
	projectAccess := &fakeProjectAccess{role: domainproject.RoleViewer}
	recorder := &recordingLabelEventRecorder{}
	service := NewService(repo, projectAccess, recorder)

	labels, err := service.ListLabels(context.Background(), ListLabelsInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
	})
	if err != nil {
		t.Fatalf("list labels: %v", err)
	}
	if len(labels) != 1 || labels[0].ID != testLabelID {
		t.Fatalf("expected labels, got %#v", labels)
	}
	if repo.gotListProjectID != testProjectID {
		t.Fatalf("expected list project id, got %q", repo.gotListProjectID)
	}
	if projectAccess.gotRoleProjectID != "" {
		t.Fatalf("expected list not to require role lookup, got %q", projectAccess.gotRoleProjectID)
	}
	assertNoLabelEvents(t, recorder)
}

func TestCreateLabelRequiresManagerAndNormalizesInput(t *testing.T) {
	repo := &fakeLabelRepository{}
	projectAccess := &fakeProjectAccess{role: domainproject.RoleEditor}
	recorder := &recordingLabelEventRecorder{}
	service := NewService(repo, projectAccess, recorder)

	label, err := service.CreateLabel(context.Background(), CreateLabelInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		Key:           " region ",
		Value:         " tokyo ",
	})
	if err != nil {
		t.Fatalf("create label: %v", err)
	}
	if label.Key != "region" || label.Value != "tokyo" {
		t.Fatalf("expected normalized label, got %#v", label)
	}
	if repo.gotCreate.Key != "region" || repo.gotCreate.Value != "tokyo" {
		t.Fatalf("expected normalized create input, got %#v", repo.gotCreate)
	}
	assertRecordedLabelEvent(t, recorder, LabelEvent{
		Name:        LabelEventCreateSuccess,
		Action:      LabelActionCreate,
		Outcome:     LabelOutcomeSuccess,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		ProjectSlug: "engineering",
		LabelID:     testLabelID,
	})
}

func TestCreateLabelRejectsViewer(t *testing.T) {
	repo := &fakeLabelRepository{}
	recorder := &recordingLabelEventRecorder{}
	service := NewService(repo, &fakeProjectAccess{role: domainproject.RoleViewer}, recorder)

	_, err := service.CreateLabel(context.Background(), CreateLabelInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		Key:           "region",
		Value:         "tokyo",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if repo.gotCreate.Key != "" {
		t.Fatalf("expected create not to be called, got %#v", repo.gotCreate)
	}
	assertRecordedLabelEvent(t, recorder, LabelEvent{
		Name:        LabelEventCreateFailure,
		Action:      LabelActionCreate,
		Outcome:     LabelOutcomeFailure,
		Reason:      LabelReasonForbidden,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		ProjectSlug: "engineering",
	})
}

func TestCreateLabelPropagatesDuplicate(t *testing.T) {
	recorder := &recordingLabelEventRecorder{}
	service := NewService(&fakeLabelRepository{createErr: ErrLabelAlreadyExists}, &fakeProjectAccess{role: domainproject.RoleOwner}, recorder)

	_, err := service.CreateLabel(context.Background(), CreateLabelInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		Key:           "region",
		Value:         "tokyo",
	})
	if !errors.Is(err, ErrLabelAlreadyExists) {
		t.Fatalf("expected duplicate label, got %v", err)
	}
	assertRecordedLabelEvent(t, recorder, LabelEvent{
		Name:        LabelEventCreateFailure,
		Action:      LabelActionCreate,
		Outcome:     LabelOutcomeFailure,
		Reason:      LabelReasonLabelAlreadyExists,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		ProjectSlug: "engineering",
	})
}

func TestUpdateLabelUsesExistingFieldsForPartialPatch(t *testing.T) {
	value := " osaka "
	repo := &fakeLabelRepository{
		label: domainlabel.Label{ID: testLabelID, ProjectID: testProjectID, Key: "region", Value: "tokyo"},
	}
	recorder := &recordingLabelEventRecorder{}
	service := NewService(repo, &fakeProjectAccess{role: domainproject.RoleAdmin}, recorder)

	label, err := service.UpdateLabel(context.Background(), UpdateLabelInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		LabelID:       testLabelID,
		Value:         &value,
	})
	if err != nil {
		t.Fatalf("update label: %v", err)
	}
	if label.Key != "region" || label.Value != "osaka" {
		t.Fatalf("expected partial update, got %#v", label)
	}
	if repo.gotUpdate.Key != "region" || repo.gotUpdate.Value != "osaka" {
		t.Fatalf("expected update input to preserve key and trim value, got %#v", repo.gotUpdate)
	}
	assertRecordedLabelEvent(t, recorder, LabelEvent{
		Name:        LabelEventUpdateSuccess,
		Action:      LabelActionUpdate,
		Outcome:     LabelOutcomeSuccess,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		ProjectSlug: "engineering",
		LabelID:     testLabelID,
	})
}

func TestUpdateLabelRejectsEmptyPatch(t *testing.T) {
	repo := &fakeLabelRepository{}
	recorder := &recordingLabelEventRecorder{}
	service := NewService(repo, &fakeProjectAccess{role: domainproject.RoleOwner}, recorder)

	_, err := service.UpdateLabel(context.Background(), UpdateLabelInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		LabelID:       testLabelID,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if repo.gotLabelID != "" {
		t.Fatalf("expected label lookup not to be called, got %q", repo.gotLabelID)
	}
	assertRecordedLabelEvent(t, recorder, LabelEvent{
		Name:        LabelEventUpdateFailure,
		Action:      LabelActionUpdate,
		Outcome:     LabelOutcomeFailure,
		Reason:      LabelReasonInvalidInput,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		ProjectSlug: "engineering",
		LabelID:     testLabelID,
	})
}

func TestDeleteLabelRequiresManager(t *testing.T) {
	repo := &fakeLabelRepository{}
	recorder := &recordingLabelEventRecorder{}
	service := NewService(repo, &fakeProjectAccess{role: domainproject.RoleAdmin}, recorder)

	err := service.DeleteLabel(context.Background(), DeleteLabelInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		LabelID:       testLabelID,
	})
	if err != nil {
		t.Fatalf("delete label: %v", err)
	}
	if repo.gotDeleteProjectID != testProjectID || repo.gotDeleteLabelID != testLabelID {
		t.Fatalf("expected delete ids, got project=%q label=%q", repo.gotDeleteProjectID, repo.gotDeleteLabelID)
	}
	assertRecordedLabelEvent(t, recorder, LabelEvent{
		Name:        LabelEventDeleteSuccess,
		Action:      LabelActionDelete,
		Outcome:     LabelOutcomeSuccess,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		ProjectSlug: "engineering",
		LabelID:     testLabelID,
	})
}

func TestResolveLabelsDelegatesToRepository(t *testing.T) {
	repo := &fakeLabelRepository{
		labels: []domainlabel.Label{{ID: testLabelID, ProjectID: testProjectID, Key: "region", Value: "tokyo"}},
	}
	recorder := &recordingLabelEventRecorder{}
	service := NewService(repo, &fakeProjectAccess{}, recorder)

	labels, err := service.GetActiveLabelsByIDsForProject(context.Background(), testProjectID, []string{testLabelID})
	if err != nil {
		t.Fatalf("resolve labels: %v", err)
	}
	if len(labels) != 1 || labels[0].ID != testLabelID {
		t.Fatalf("expected resolved label, got %#v", labels)
	}
	if repo.gotResolveProjectID != testProjectID || len(repo.gotResolveLabelIDs) != 1 || repo.gotResolveLabelIDs[0] != testLabelID {
		t.Fatalf("expected resolve inputs, got project=%q labels=%#v", repo.gotResolveProjectID, repo.gotResolveLabelIDs)
	}
	assertNoLabelEvents(t, recorder)
}

func TestListLabelsRecordsTechnicalFailure(t *testing.T) {
	recorder := &recordingLabelEventRecorder{}
	listErr := errors.New("list labels")
	service := NewService(&fakeLabelRepository{listErr: listErr}, &fakeProjectAccess{}, recorder)

	_, err := service.ListLabels(context.Background(), ListLabelsInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
	})
	if !errors.Is(err, listErr) {
		t.Fatalf("expected list error, got %v", err)
	}
	assertRecordedLabelEvent(t, recorder, LabelEvent{
		Name:        LabelEventListFailure,
		Action:      LabelActionList,
		Outcome:     LabelOutcomeFailure,
		Reason:      LabelReasonLabelListFailed,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		ProjectSlug: "engineering",
		Err:         listErr,
	})
}

func TestCreateLabelRecordsRoleLookupFailure(t *testing.T) {
	recorder := &recordingLabelEventRecorder{}
	roleErr := errors.New("lookup role")
	service := NewService(&fakeLabelRepository{}, &fakeProjectAccess{roleErr: roleErr}, recorder)

	_, err := service.CreateLabel(context.Background(), CreateLabelInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		Key:           "region",
		Value:         "tokyo",
	})
	if !errors.Is(err, roleErr) {
		t.Fatalf("expected role lookup error, got %v", err)
	}
	assertRecordedLabelEvent(t, recorder, LabelEvent{
		Name:        LabelEventCreateFailure,
		Action:      LabelActionCreate,
		Outcome:     LabelOutcomeFailure,
		Reason:      LabelReasonRoleLookupFailed,
		ActorUserID: "user-1",
		ProjectID:   testProjectID,
		ProjectRef:  "engineering",
		ProjectSlug: "engineering",
		Err:         roleErr,
	})
}

func TestResolveLabelsRecordsFailure(t *testing.T) {
	recorder := &recordingLabelEventRecorder{}
	service := NewService(&fakeLabelRepository{resolveErr: ErrLabelNotFound}, &fakeProjectAccess{}, recorder)

	_, err := service.GetActiveLabelsByIDsForProject(context.Background(), testProjectID, []string{testLabelID})
	if !errors.Is(err, ErrLabelNotFound) {
		t.Fatalf("expected label not found, got %v", err)
	}
	assertRecordedLabelEvent(t, recorder, LabelEvent{
		Name:      LabelEventResolveFailure,
		Action:    LabelActionResolve,
		Outcome:   LabelOutcomeFailure,
		Reason:    LabelReasonLabelNotFound,
		ProjectID: testProjectID,
	})
}

func assertRecordedLabelEvent(t *testing.T, recorder *recordingLabelEventRecorder, want LabelEvent) {
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
		got.LabelID != want.LabelID ||
		!errors.Is(got.Err, want.Err) {
		t.Fatalf("unexpected event:\n got: %#v\nwant: %#v", got, want)
	}
}

func assertNoLabelEvents(t *testing.T, recorder *recordingLabelEventRecorder) {
	t.Helper()

	if len(recorder.events) != 0 {
		t.Fatalf("expected no events, got %d: %#v", len(recorder.events), recorder.events)
	}
}

type recordingLabelEventRecorder struct {
	events []LabelEvent
}

func (r *recordingLabelEventRecorder) RecordLabelEvent(_ context.Context, event LabelEvent) {
	r.events = append(r.events, event)
}

type fakeLabelRepository struct {
	labels              []domainlabel.Label
	label               domainlabel.Label
	gotListProjectID    string
	listErr             error
	gotLabelID          string
	gotCreate           domainlabel.CreateLabelStorageInput
	createErr           error
	gotUpdate           domainlabel.UpdateLabelStorageInput
	updateErr           error
	gotDeleteProjectID  string
	gotDeleteLabelID    string
	deleteErr           error
	gotResolveProjectID string
	gotResolveLabelIDs  []string
	resolveErr          error
}

func (r *fakeLabelRepository) ListLabels(_ context.Context, projectID string) ([]domainlabel.Label, error) {
	r.gotListProjectID = projectID
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.labels, nil
}

func (r *fakeLabelRepository) GetLabel(_ context.Context, projectID, labelID string) (domainlabel.Label, error) {
	r.gotListProjectID = projectID
	r.gotLabelID = labelID
	if r.label.ID != "" {
		return r.label, nil
	}
	return domainlabel.Label{ID: labelID, ProjectID: projectID, Key: "region", Value: "tokyo"}, nil
}

func (r *fakeLabelRepository) CreateLabel(_ context.Context, input domainlabel.CreateLabelStorageInput) (domainlabel.Label, error) {
	r.gotCreate = input
	if r.createErr != nil {
		return domainlabel.Label{}, r.createErr
	}
	return newFakeLabel(input.ProjectID, input.Key, input.Value), nil
}

func (r *fakeLabelRepository) UpdateLabel(_ context.Context, input domainlabel.UpdateLabelStorageInput) (domainlabel.Label, error) {
	r.gotUpdate = input
	if r.updateErr != nil {
		return domainlabel.Label{}, r.updateErr
	}
	return domainlabel.Label{ID: input.LabelID, ProjectID: input.ProjectID, Key: input.Key, Value: input.Value, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (r *fakeLabelRepository) SoftDeleteLabel(_ context.Context, projectID, labelID string) error {
	r.gotDeleteProjectID = projectID
	r.gotDeleteLabelID = labelID
	return r.deleteErr
}

func (r *fakeLabelRepository) GetActiveLabelsByIDsForProject(_ context.Context, projectID string, labelIDs []string) ([]domainlabel.Label, error) {
	r.gotResolveProjectID = projectID
	r.gotResolveLabelIDs = append([]string(nil), labelIDs...)
	if r.resolveErr != nil {
		return nil, r.resolveErr
	}
	return r.labels, nil
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

func newFakeLabel(projectID, key, value string) domainlabel.Label {
	return domainlabel.Label{
		ID:        testLabelID,
		ProjectID: projectID,
		Key:       key,
		Value:     value,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
