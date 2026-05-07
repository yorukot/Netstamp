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
	service := NewService(repo)

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
	service := NewService(repo)

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
}

func TestDeleteProjectRequiresOwner(t *testing.T) {
	repo := &fakeProjectRepository{actorRole: domainproject.RoleAdmin}
	service := NewService(repo)

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
			service := NewService(repo)

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
			service := NewService(repo)

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

type fakeProjectRepository struct {
	createdProject         domainproject.Project
	gotCreateInput         CreateProjectStorageInput
	projects               []domainproject.Project
	project                domainproject.Project
	actorRole              domainproject.Role
	roleErr                error
	updatedProject         domainproject.Project
	gotUpdateProject       UpdateProjectStorageInput
	gotSoftDeleteProjectID string
	members                []domainproject.Member
	member                 domainproject.Member
	addedMember            domainproject.Member
	gotAddMember           AddMemberStorageInput
	updatedMember          domainproject.Member
	gotUpdateMemberRole    UpdateMemberRoleStorageInput
	owners                 int
}

func (r *fakeProjectRepository) CreateProjectWithOwner(_ context.Context, input CreateProjectStorageInput) (domainproject.Project, error) {
	r.gotCreateInput = input
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
	return r.projects, nil
}

func (r *fakeProjectRepository) GetProjectForUser(context.Context, string, string) (domainproject.Project, error) {
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

func (r *fakeProjectRepository) UpdateProject(_ context.Context, input UpdateProjectStorageInput) (domainproject.Project, error) {
	r.gotUpdateProject = input
	if r.updatedProject.ID != "" {
		return r.updatedProject, nil
	}
	return domainproject.Project{ID: input.ProjectID, Name: input.Name, Slug: input.Slug}, nil
}

func (r *fakeProjectRepository) SoftDeleteProject(_ context.Context, projectID string) error {
	r.gotSoftDeleteProjectID = projectID
	return nil
}

func (r *fakeProjectRepository) ListMembers(context.Context, string) ([]domainproject.Member, error) {
	return r.members, nil
}

func (r *fakeProjectRepository) GetMember(context.Context, string, string) (domainproject.Member, error) {
	if r.member.ID != "" {
		return r.member, nil
	}
	return domainproject.Member{ID: "member-1", UserID: "target-user", Role: domainproject.RoleViewer}, nil
}

func (r *fakeProjectRepository) AddMember(_ context.Context, input AddMemberStorageInput) (domainproject.Member, error) {
	r.gotAddMember = input
	if r.addedMember.ID != "" {
		return r.addedMember, nil
	}
	return domainproject.Member{ID: "member-1", ProjectID: input.ProjectID, UserID: input.UserID, Role: input.Role}, nil
}

func (r *fakeProjectRepository) UpdateMemberRole(_ context.Context, input UpdateMemberRoleStorageInput) (domainproject.Member, error) {
	r.gotUpdateMemberRole = input
	if r.updatedMember.ID != "" {
		return r.updatedMember, nil
	}
	return domainproject.Member{ID: "member-1", ProjectID: input.ProjectID, UserID: input.UserID, Role: input.Role}, nil
}

func (r *fakeProjectRepository) CountOwners(context.Context, string) (int, error) {
	if r.owners > 0 {
		return r.owners, nil
	}
	return 1, nil
}
