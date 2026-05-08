package project

import (
	"context"
	"errors"
	"testing"
	"time"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func TestCreateProjectNormalizesInputAndCreatesOwnerMembership(t *testing.T) {
	repo := &fakeProjectRepository{
		createdProject: domainproject.Project{ID: "project-1"},
	}
	service := NewService(repo, nil)

	_, err := service.CreateProject(context.Background(), CreateProjectInput{
		CurrentUserID: "user-1",
		Name:          "  Engineering  ",
		Slug:          "  platform-project  ",
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	if repo.gotCreateInput.Name != "Engineering" {
		t.Fatalf("expected trimmed name, got %q", repo.gotCreateInput.Name)
	}
	if repo.gotCreateInput.Slug != "platform-project" {
		t.Fatalf("expected trimmed slug, got %q", repo.gotCreateInput.Slug)
	}
	if repo.gotCreateInput.CreatedByUserID != "user-1" {
		t.Fatalf("expected owner user id, got %q", repo.gotCreateInput.CreatedByUserID)
	}
}

func TestCreateProjectRejectsInvalidSlug(t *testing.T) {
	repo := &fakeProjectRepository{}
	recorder := &recordingProjectEventRecorder{}
	service := NewService(repo, recorder)

	_, err := service.CreateProject(context.Background(), CreateProjectInput{
		CurrentUserID: "user-1",
		Name:          "Engineering",
		Slug:          "Platform_Project",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	if repo.gotCreateInput.Slug != "" {
		t.Fatalf("expected create not to be called, got %#v", repo.gotCreateInput)
	}
	assertRecordedProjectEvent(t, recorder, ProjectEvent{
		Name:        ProjectEventCreateFailure,
		Action:      ProjectActionCreate,
		Outcome:     ProjectOutcomeFailure,
		Reason:      ProjectReasonInvalidInput,
		ActorUserID: "user-1",
	})
}

func TestCreateProjectRecordsSuccess(t *testing.T) {
	recorder := &recordingProjectEventRecorder{}
	repo := &fakeProjectRepository{
		createdProject: domainproject.Project{ID: "project-1", Slug: "engineering"},
	}
	service := NewService(repo, recorder)

	_, err := service.CreateProject(context.Background(), CreateProjectInput{
		CurrentUserID: "user-1",
		Name:          "Engineering",
		Slug:          "engineering",
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	assertRecordedProjectEvent(t, recorder, ProjectEvent{
		Name:        ProjectEventCreateSuccess,
		Action:      ProjectActionCreate,
		Outcome:     ProjectOutcomeSuccess,
		ActorUserID: "user-1",
		ProjectID:   "project-1",
		ProjectSlug: "engineering",
	})
}

func TestCreateProjectRecordsTechnicalFailure(t *testing.T) {
	recorder := &recordingProjectEventRecorder{}
	createErr := errors.New("insert project")
	repo := &fakeProjectRepository{createErr: createErr}
	service := NewService(repo, recorder)

	_, err := service.CreateProject(context.Background(), CreateProjectInput{
		CurrentUserID: "user-1",
		Name:          "Engineering",
		Slug:          "engineering",
	})
	if !errors.Is(err, createErr) {
		t.Fatalf("expected create error, got %v", err)
	}

	assertRecordedProjectEvent(t, recorder, ProjectEvent{
		Name:        ProjectEventCreateFailure,
		Action:      ProjectActionCreate,
		Outcome:     ProjectOutcomeFailure,
		Reason:      ProjectReasonProjectCreateFailed,
		ActorUserID: "user-1",
		ProjectSlug: "engineering",
		Err:         createErr,
	})
}

func TestDeleteProjectRequiresOwner(t *testing.T) {
	repo := &fakeProjectRepository{actorRole: domainproject.RoleAdmin}
	recorder := &recordingProjectEventRecorder{}
	service := NewService(repo, recorder)

	err := service.DeleteProject(context.Background(), DeleteProjectInput{
		CurrentUserID: "admin-user",
		ProjectRef:    "project-1",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if repo.gotSoftDeleteProjectID != "" {
		t.Fatalf("expected delete not to be called, got %q", repo.gotSoftDeleteProjectID)
	}
	assertRecordedProjectEvent(t, recorder, ProjectEvent{
		Name:        ProjectEventDeleteFailure,
		Action:      ProjectActionDelete,
		Outcome:     ProjectOutcomeFailure,
		Reason:      ProjectReasonForbidden,
		ActorUserID: "admin-user",
		ProjectID:   "project-1",
		ProjectRef:  "project-1",
		ProjectSlug: "engineering",
	})
}

func TestDeleteProjectRecordsSuccess(t *testing.T) {
	recorder := &recordingProjectEventRecorder{}
	repo := &fakeProjectRepository{actorRole: domainproject.RoleOwner}
	service := NewService(repo, recorder)

	err := service.DeleteProject(context.Background(), DeleteProjectInput{
		CurrentUserID: "owner-user",
		ProjectRef:    "engineering",
	})
	if err != nil {
		t.Fatalf("delete project: %v", err)
	}

	assertRecordedProjectEvent(t, recorder, ProjectEvent{
		Name:        ProjectEventDeleteSuccess,
		Action:      ProjectActionDelete,
		Outcome:     ProjectOutcomeSuccess,
		ActorUserID: "owner-user",
		ProjectID:   "project-1",
		ProjectRef:  "engineering",
		ProjectSlug: "engineering",
	})
}

func TestAddMemberRoleRestrictions(t *testing.T) {
	tests := []struct {
		name      string
		actorRole domainproject.Role
		newRole   domainproject.Role
		wantErr   error
	}{
		{
			name:      "owner cannot add owner",
			actorRole: domainproject.RoleOwner,
			newRole:   domainproject.RoleOwner,
			wantErr:   ErrForbidden,
		},
		{
			name:      "admin cannot add admin",
			actorRole: domainproject.RoleAdmin,
			newRole:   domainproject.RoleAdmin,
			wantErr:   ErrForbidden,
		},
		{
			name:      "owner can add admin",
			actorRole: domainproject.RoleOwner,
			newRole:   domainproject.RoleAdmin,
		},
		{
			name:      "admin can add viewer",
			actorRole: domainproject.RoleAdmin,
			newRole:   domainproject.RoleViewer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeProjectRepository{
				actorRole:   tt.actorRole,
				addedMember: domainproject.Member{ID: "member-1", Role: tt.newRole},
			}
			service := NewService(repo, nil)

			_, err := service.AddMember(context.Background(), AddMemberInput{
				CurrentUserID: "actor-user",
				ProjectRef:    "project-1",
				UserID:        "target-user",
				Role:          tt.newRole,
			})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && repo.gotAddMember.UserID != "" {
				t.Fatalf("expected add not to be called, got %#v", repo.gotAddMember)
			}
			if tt.wantErr == nil && repo.gotAddMember.Role != tt.newRole {
				t.Fatalf("expected add role %q, got %q", tt.newRole, repo.gotAddMember.Role)
			}
		})
	}
}

func TestUpdateMemberRoleRestrictions(t *testing.T) {
	tests := []struct {
		name       string
		actorRole  domainproject.Role
		memberRole domainproject.Role
		newRole    domainproject.Role
		owners     int
		wantErr    error
	}{
		{
			name:       "owner cannot change anyone to owner",
			actorRole:  domainproject.RoleOwner,
			memberRole: domainproject.RoleViewer,
			newRole:    domainproject.RoleOwner,
			owners:     1,
			wantErr:    ErrForbidden,
		},
		{
			name:       "admin cannot change anyone to admin",
			actorRole:  domainproject.RoleAdmin,
			memberRole: domainproject.RoleViewer,
			newRole:    domainproject.RoleAdmin,
			owners:     1,
			wantErr:    ErrForbidden,
		},
		{
			name:       "admin cannot change owner",
			actorRole:  domainproject.RoleAdmin,
			memberRole: domainproject.RoleOwner,
			newRole:    domainproject.RoleViewer,
			owners:     2,
			wantErr:    ErrForbidden,
		},
		{
			name:       "cannot remove last owner",
			actorRole:  domainproject.RoleOwner,
			memberRole: domainproject.RoleOwner,
			newRole:    domainproject.RoleAdmin,
			owners:     1,
			wantErr:    ErrLastOwner,
		},
		{
			name:       "owner can change member to admin",
			actorRole:  domainproject.RoleOwner,
			memberRole: domainproject.RoleViewer,
			newRole:    domainproject.RoleAdmin,
			owners:     1,
		},
		{
			name:       "admin can change member to editor",
			actorRole:  domainproject.RoleAdmin,
			memberRole: domainproject.RoleViewer,
			newRole:    domainproject.RoleEditor,
			owners:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeProjectRepository{
				actorRole: tt.actorRole,
				member: domainproject.Member{
					ID:     "member-1",
					UserID: "target-user",
					Role:   tt.memberRole,
				},
				owners:        tt.owners,
				updatedMember: domainproject.Member{ID: "member-1", Role: tt.newRole},
			}
			service := NewService(repo, nil)

			_, err := service.UpdateMemberRole(context.Background(), UpdateMemberRoleInput{
				CurrentUserID: "actor-user",
				ProjectRef:    "project-1",
				UserID:        "target-user",
				Role:          tt.newRole,
			})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil && repo.gotUpdateMemberRole.UserID != "" {
				t.Fatalf("expected update not to be called, got %#v", repo.gotUpdateMemberRole)
			}
			if tt.wantErr == nil && repo.gotUpdateMemberRole.Role != tt.newRole {
				t.Fatalf("expected update role %q, got %q", tt.newRole, repo.gotUpdateMemberRole.Role)
			}
		})
	}
}

func TestAddMemberRecordsSuccess(t *testing.T) {
	recorder := &recordingProjectEventRecorder{}
	repo := &fakeProjectRepository{
		actorRole:   domainproject.RoleOwner,
		addedMember: domainproject.Member{ID: "member-1", UserID: "target-user", Role: domainproject.RoleAdmin},
	}
	service := NewService(repo, recorder)

	_, err := service.AddMember(context.Background(), AddMemberInput{
		CurrentUserID: "actor-user",
		ProjectRef:    "engineering",
		UserID:        "target-user",
		Role:          domainproject.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("add member: %v", err)
	}

	assertRecordedProjectEvent(t, recorder, ProjectEvent{
		Name:         ProjectEventAddMemberSuccess,
		Action:       ProjectActionAddMember,
		Outcome:      ProjectOutcomeSuccess,
		ActorUserID:  "actor-user",
		ProjectID:    "project-1",
		ProjectRef:   "engineering",
		ProjectSlug:  "engineering",
		TargetUserID: "target-user",
		Role:         domainproject.RoleAdmin,
	})
}

func TestUpdateProjectSuccessDoesNotRecordEvent(t *testing.T) {
	recorder := &recordingProjectEventRecorder{}
	repo := &fakeProjectRepository{
		actorRole:      domainproject.RoleOwner,
		updatedProject: domainproject.Project{ID: "project-1", Name: "Platform", Slug: "engineering"},
	}
	service := NewService(repo, recorder)
	name := "Platform"

	_, err := service.UpdateProject(context.Background(), UpdateProjectInput{
		CurrentUserID: "actor-user",
		ProjectRef:    "engineering",
		Name:          &name,
	})
	if err != nil {
		t.Fatalf("update project: %v", err)
	}

	assertNoProjectEvents(t, recorder)
}

func TestReadSuccessDoesNotRecordEvent(t *testing.T) {
	recorder := &recordingProjectEventRecorder{}
	repo := &fakeProjectRepository{}
	service := NewService(repo, recorder)

	if _, err := service.ListProjects(context.Background(), ListProjectsInput{CurrentUserID: "user-1"}); err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if _, err := service.GetProject(context.Background(), GetProjectInput{CurrentUserID: "user-1", ProjectRef: "engineering"}); err != nil {
		t.Fatalf("get project: %v", err)
	}
	if _, err := service.ListMembers(context.Background(), ListMembersInput{CurrentUserID: "user-1", ProjectRef: "engineering"}); err != nil {
		t.Fatalf("list members: %v", err)
	}

	assertNoProjectEvents(t, recorder)
}

func TestReadBusinessFailureDoesNotRecordEvent(t *testing.T) {
	recorder := &recordingProjectEventRecorder{}
	repo := &fakeProjectRepository{getErr: ErrProjectNotFound}
	service := NewService(repo, recorder)

	_, err := service.GetProject(context.Background(), GetProjectInput{
		CurrentUserID: "user-1",
		ProjectRef:    "missing",
	})
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("expected project not found, got %v", err)
	}

	assertNoProjectEvents(t, recorder)
}

func TestReadTechnicalFailureRecordsEvent(t *testing.T) {
	recorder := &recordingProjectEventRecorder{}
	listErr := errors.New("list projects")
	repo := &fakeProjectRepository{listErr: listErr}
	service := NewService(repo, recorder)

	_, err := service.ListProjects(context.Background(), ListProjectsInput{CurrentUserID: "user-1"})
	if !errors.Is(err, listErr) {
		t.Fatalf("expected list error, got %v", err)
	}

	assertRecordedProjectEvent(t, recorder, ProjectEvent{
		Name:        ProjectEventListFailure,
		Action:      ProjectActionList,
		Outcome:     ProjectOutcomeFailure,
		Reason:      ProjectReasonProjectListFailed,
		ActorUserID: "user-1",
		Err:         listErr,
	})
}

func TestUpdateMemberRoleRecordsSuccess(t *testing.T) {
	recorder := &recordingProjectEventRecorder{}
	repo := &fakeProjectRepository{
		actorRole:     domainproject.RoleOwner,
		member:        domainproject.Member{ID: "member-1", UserID: "target-user", Role: domainproject.RoleViewer},
		updatedMember: domainproject.Member{ID: "member-1", UserID: "target-user", Role: domainproject.RoleAdmin},
	}
	service := NewService(repo, recorder)

	_, err := service.UpdateMemberRole(context.Background(), UpdateMemberRoleInput{
		CurrentUserID: "actor-user",
		ProjectRef:    "engineering",
		UserID:        "target-user",
		Role:          domainproject.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("update member role: %v", err)
	}

	assertRecordedProjectEvent(t, recorder, ProjectEvent{
		Name:         ProjectEventUpdateMemberRoleSuccess,
		Action:       ProjectActionUpdateMemberRole,
		Outcome:      ProjectOutcomeSuccess,
		ActorUserID:  "actor-user",
		ProjectID:    "project-1",
		ProjectRef:   "engineering",
		ProjectSlug:  "engineering",
		TargetUserID: "target-user",
		Role:         domainproject.RoleAdmin,
	})
}

func TestUpdateMemberRoleLastOwnerRecordsBusinessFailure(t *testing.T) {
	recorder := &recordingProjectEventRecorder{}
	repo := &fakeProjectRepository{
		actorRole: domainproject.RoleOwner,
		member:    domainproject.Member{ID: "member-1", UserID: "target-user", Role: domainproject.RoleOwner},
		owners:    1,
	}
	service := NewService(repo, recorder)

	_, err := service.UpdateMemberRole(context.Background(), UpdateMemberRoleInput{
		CurrentUserID: "actor-user",
		ProjectRef:    "engineering",
		UserID:        "target-user",
		Role:          domainproject.RoleAdmin,
	})
	if !errors.Is(err, ErrLastOwner) {
		t.Fatalf("expected last owner, got %v", err)
	}

	assertRecordedProjectEvent(t, recorder, ProjectEvent{
		Name:         ProjectEventUpdateMemberRoleFailure,
		Action:       ProjectActionUpdateMemberRole,
		Outcome:      ProjectOutcomeFailure,
		Reason:       ProjectReasonLastOwner,
		ActorUserID:  "actor-user",
		ProjectID:    "project-1",
		ProjectRef:   "engineering",
		ProjectSlug:  "engineering",
		TargetUserID: "target-user",
		Role:         domainproject.RoleAdmin,
	})
}

func assertRecordedProjectEvent(t *testing.T, recorder *recordingProjectEventRecorder, want ProjectEvent) {
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
		got.TargetUserID != want.TargetUserID ||
		got.Role != want.Role ||
		!errors.Is(got.Err, want.Err) {
		t.Fatalf("unexpected event:\n got: %#v\nwant: %#v", got, want)
	}
}

func assertNoProjectEvents(t *testing.T, recorder *recordingProjectEventRecorder) {
	t.Helper()

	if len(recorder.events) != 0 {
		t.Fatalf("expected no events, got %d: %#v", len(recorder.events), recorder.events)
	}
}

type recordingProjectEventRecorder struct {
	events []ProjectEvent
}

func (r *recordingProjectEventRecorder) RecordProjectEvent(_ context.Context, event ProjectEvent) {
	r.events = append(r.events, event)
}

type fakeProjectRepository struct {
	createdProject         domainproject.Project
	createErr              error
	gotCreateInput         domainproject.CreateProjectStorageInput
	projects               []domainproject.Project
	listErr                error
	project                domainproject.Project
	getErr                 error
	actorRole              domainproject.Role
	roleErr                error
	updatedProject         domainproject.Project
	updateErr              error
	gotUpdateProject       domainproject.UpdateProjectStorageInput
	deleteErr              error
	gotSoftDeleteProjectID string
	members                []domainproject.Member
	listMembersErr         error
	member                 domainproject.Member
	getMemberErr           error
	addedMember            domainproject.Member
	addMemberErr           error
	gotAddMember           domainproject.AddMemberStorageInput
	updatedMember          domainproject.Member
	updateMemberErr        error
	gotUpdateMemberRole    domainproject.UpdateMemberRoleStorageInput
	owners                 int
	countOwnersErr         error
}

func (r *fakeProjectRepository) CreateProjectWithOwner(_ context.Context, input domainproject.CreateProjectStorageInput) (domainproject.Project, error) {
	r.gotCreateInput = input
	if r.createErr != nil {
		return domainproject.Project{}, r.createErr
	}
	if r.createdProject.ID != "" {
		return r.createdProject, nil
	}
	return domainproject.Project{
		ID:              "project-1",
		Name:            input.Name,
		Slug:            input.Slug,
		CreatedByUserID: input.CreatedByUserID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

func (r *fakeProjectRepository) ListProjectsForUser(context.Context, string) ([]domainproject.Project, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.projects, nil
}

func (r *fakeProjectRepository) GetProjectForUser(context.Context, string, string) (domainproject.Project, error) {
	if r.getErr != nil {
		return domainproject.Project{}, r.getErr
	}
	if r.project.ID != "" {
		return r.project, nil
	}
	return domainproject.Project{ID: "project-1", Name: "Engineering", Slug: "engineering"}, nil
}

func (r *fakeProjectRepository) GetMemberRole(context.Context, string, string) (domainproject.Role, error) {
	if r.roleErr != nil {
		return "", r.roleErr
	}
	if r.actorRole != "" {
		return r.actorRole, nil
	}
	return domainproject.RoleOwner, nil
}

func (r *fakeProjectRepository) UpdateProject(_ context.Context, input domainproject.UpdateProjectStorageInput) (domainproject.Project, error) {
	r.gotUpdateProject = input
	if r.updateErr != nil {
		return domainproject.Project{}, r.updateErr
	}
	if r.updatedProject.ID != "" {
		return r.updatedProject, nil
	}
	return domainproject.Project{ID: input.ProjectID, Name: input.Name, Slug: input.Slug}, nil
}

func (r *fakeProjectRepository) SoftDeleteProject(_ context.Context, projectID string) error {
	r.gotSoftDeleteProjectID = projectID
	return r.deleteErr
}

func (r *fakeProjectRepository) ListMembers(context.Context, string) ([]domainproject.Member, error) {
	if r.listMembersErr != nil {
		return nil, r.listMembersErr
	}
	return r.members, nil
}

func (r *fakeProjectRepository) GetMember(context.Context, string, string) (domainproject.Member, error) {
	if r.getMemberErr != nil {
		return domainproject.Member{}, r.getMemberErr
	}
	if r.member.ID != "" {
		return r.member, nil
	}
	return domainproject.Member{ID: "member-1", UserID: "target-user", Role: domainproject.RoleViewer}, nil
}

func (r *fakeProjectRepository) AddMember(_ context.Context, input domainproject.AddMemberStorageInput) (domainproject.Member, error) {
	r.gotAddMember = input
	if r.addMemberErr != nil {
		return domainproject.Member{}, r.addMemberErr
	}
	if r.addedMember.ID != "" {
		return r.addedMember, nil
	}
	return domainproject.Member{ID: "member-1", ProjectID: input.ProjectID, UserID: input.UserID, Role: input.Role}, nil
}

func (r *fakeProjectRepository) UpdateMemberRole(_ context.Context, input domainproject.UpdateMemberRoleStorageInput) (domainproject.Member, error) {
	r.gotUpdateMemberRole = input
	if r.updateMemberErr != nil {
		return domainproject.Member{}, r.updateMemberErr
	}
	if r.updatedMember.ID != "" {
		return r.updatedMember, nil
	}
	return domainproject.Member{ID: "member-1", ProjectID: input.ProjectID, UserID: input.UserID, Role: input.Role}, nil
}

func (r *fakeProjectRepository) CountOwners(context.Context, string) (int, error) {
	if r.countOwnersErr != nil {
		return 0, r.countOwnersErr
	}
	if r.owners > 0 {
		return r.owners, nil
	}
	return 1, nil
}
