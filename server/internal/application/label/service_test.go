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
	service := NewService(repo, projectAccess)

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
}

func TestCreateLabelRequiresManagerAndNormalizesInput(t *testing.T) {
	repo := &fakeLabelRepository{}
	projectAccess := &fakeProjectAccess{role: domainproject.RoleEditor}
	service := NewService(repo, projectAccess)

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
}

func TestCreateLabelRejectsViewer(t *testing.T) {
	repo := &fakeLabelRepository{}
	service := NewService(repo, &fakeProjectAccess{role: domainproject.RoleViewer})

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
}

func TestCreateLabelPropagatesDuplicate(t *testing.T) {
	service := NewService(&fakeLabelRepository{createErr: ErrLabelAlreadyExists}, &fakeProjectAccess{role: domainproject.RoleOwner})

	_, err := service.CreateLabel(context.Background(), CreateLabelInput{
		CurrentUserID: "user-1",
		ProjectRef:    "engineering",
		Key:           "region",
		Value:         "tokyo",
	})
	if !errors.Is(err, ErrLabelAlreadyExists) {
		t.Fatalf("expected duplicate label, got %v", err)
	}
}

func TestUpdateLabelUsesExistingFieldsForPartialPatch(t *testing.T) {
	value := " osaka "
	repo := &fakeLabelRepository{
		label: domainlabel.Label{ID: testLabelID, ProjectID: testProjectID, Key: "region", Value: "tokyo"},
	}
	service := NewService(repo, &fakeProjectAccess{role: domainproject.RoleAdmin})

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
}

func TestUpdateLabelRejectsEmptyPatch(t *testing.T) {
	repo := &fakeLabelRepository{}
	service := NewService(repo, &fakeProjectAccess{role: domainproject.RoleOwner})

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
}

func TestDeleteLabelRequiresManager(t *testing.T) {
	repo := &fakeLabelRepository{}
	service := NewService(repo, &fakeProjectAccess{role: domainproject.RoleAdmin})

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
}

func TestResolveLabelsDelegatesToRepository(t *testing.T) {
	repo := &fakeLabelRepository{
		labels: []domainlabel.Label{{ID: testLabelID, ProjectID: testProjectID, Key: "region", Value: "tokyo"}},
	}
	service := NewService(repo, &fakeProjectAccess{})

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
}

type fakeLabelRepository struct {
	labels              []domainlabel.Label
	label               domainlabel.Label
	gotListProjectID    string
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
	return r.labels, nil
}

func (r *fakeLabelRepository) GetLabel(_ context.Context, projectID string, labelID string) (domainlabel.Label, error) {
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

func (r *fakeLabelRepository) SoftDeleteLabel(_ context.Context, projectID string, labelID string) error {
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

func (r *fakeProjectAccess) GetProjectForUser(_ context.Context, projectRef string, userID string) (domainproject.Project, error) {
	r.gotProjectRef = projectRef
	r.gotUserID = userID
	if r.projectErr != nil {
		return domainproject.Project{}, r.projectErr
	}
	return domainproject.Project{ID: testProjectID, Slug: "engineering"}, nil
}

func (r *fakeProjectAccess) GetMemberRole(_ context.Context, projectID string, userID string) (domainproject.Role, error) {
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

func newFakeLabel(projectID string, key string, value string) domainlabel.Label {
	return domainlabel.Label{
		ID:        testLabelID,
		ProjectID: projectID,
		Key:       key,
		Value:     value,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
