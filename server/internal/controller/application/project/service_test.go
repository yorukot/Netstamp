package project

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

const (
	testProjectID    = "33333333-3333-3333-3333-333333333333"
	testOtherOwnerID = "44444444-4444-4444-4444-444444444444"
	testInviteID     = "77777777-7777-7777-7777-777777777777"
)

func TestCreateInviteCreatesPendingInvite(t *testing.T) {
	repo := newProjectServiceRepository()
	repo.members[testCurrentUserID] = domainproject.Member{
		ID:        "55555555-5555-5555-5555-555555555555",
		ProjectID: testProjectID,
		UserID:    testCurrentUserID,
		Role:      domainproject.RoleOwner,
	}
	users := &projectServiceUserLookup{
		user: identity.User{
			ID:          testMemberUserID,
			Email:       "member@example.com",
			DisplayName: "Member",
		},
	}
	service := NewService(repo, users, &recordingProjectEventRecorder{})

	invite, err := service.CreateInvite(context.Background(), CreateInviteInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "project",
		Email:         "member@example.com",
		Role:          domainproject.RoleViewer,
	})
	if err != nil {
		t.Fatalf("expected invite creation to succeed: %v", err)
	}
	if invite.Status != domainproject.InviteStatusPending {
		t.Fatalf("expected pending invite, got %q", invite.Status)
	}
	if repo.createdInvite.InvitedUserID != testMemberUserID {
		t.Fatalf("expected invite target %q, got %q", testMemberUserID, repo.createdInvite.InvitedUserID)
	}
	if repo.createdInvite.InvitedEmail != "member@example.com" {
		t.Fatalf("expected invited email %q, got %q", "member@example.com", repo.createdInvite.InvitedEmail)
	}
}

func TestCreateInviteAllowsUnknownEmail(t *testing.T) {
	repo := newProjectServiceRepository()
	repo.members[testCurrentUserID] = domainproject.Member{
		ID:        "55555555-5555-5555-5555-555555555555",
		ProjectID: testProjectID,
		UserID:    testCurrentUserID,
		Role:      domainproject.RoleOwner,
	}
	users := &projectServiceUserLookup{err: identity.ErrUserNotFound}
	service := NewService(repo, users, &recordingProjectEventRecorder{})

	invite, err := service.CreateInvite(context.Background(), CreateInviteInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "project",
		Email:         "new.member@example.com",
		Role:          domainproject.RoleViewer,
	})
	if err != nil {
		t.Fatalf("expected invite creation to succeed: %v", err)
	}
	if invite.Status != domainproject.InviteStatusPending {
		t.Fatalf("expected pending invite, got %q", invite.Status)
	}
	if repo.createdInvite.InvitedEmail != "new.member@example.com" {
		t.Fatalf("expected invited email %q, got %q", "new.member@example.com", repo.createdInvite.InvitedEmail)
	}
	if repo.createdInvite.InvitedUserID != "" {
		t.Fatalf("expected unknown invite to have no user ID, got %q", repo.createdInvite.InvitedUserID)
	}
}

func TestCreateInviteRejectsExistingMember(t *testing.T) {
	repo := newProjectServiceRepository()
	repo.members[testCurrentUserID] = domainproject.Member{
		ID:        "55555555-5555-5555-5555-555555555555",
		ProjectID: testProjectID,
		UserID:    testCurrentUserID,
		Role:      domainproject.RoleOwner,
	}
	repo.members[testMemberUserID] = domainproject.Member{
		ID:        "66666666-6666-6666-6666-666666666666",
		ProjectID: testProjectID,
		UserID:    testMemberUserID,
		Role:      domainproject.RoleViewer,
	}
	users := &projectServiceUserLookup{
		user: identity.User{
			ID:          testMemberUserID,
			Email:       "member@example.com",
			DisplayName: "Member",
		},
	}
	service := NewService(repo, users, &recordingProjectEventRecorder{})

	_, err := service.CreateInvite(context.Background(), CreateInviteInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "project",
		Email:         "member@example.com",
		Role:          domainproject.RoleViewer,
	})
	if !errors.Is(err, domainproject.ErrMemberAlreadyExists) {
		t.Fatalf("expected member already exists, got %v", err)
	}
	if repo.createdInvite.ID != "" {
		t.Fatalf("expected no invite to be created, got %q", repo.createdInvite.ID)
	}
}

func TestCreateInviteRejectsUnassignableRole(t *testing.T) {
	repo := newProjectServiceRepository()
	repo.members[testCurrentUserID] = domainproject.Member{
		ID:        "55555555-5555-5555-5555-555555555555",
		ProjectID: testProjectID,
		UserID:    testCurrentUserID,
		Role:      domainproject.RoleAdmin,
	}
	users := &projectServiceUserLookup{
		user: identity.User{
			ID:          testMemberUserID,
			Email:       "member@example.com",
			DisplayName: "Member",
		},
	}
	service := NewService(repo, users, &recordingProjectEventRecorder{})

	_, err := service.CreateInvite(context.Background(), CreateInviteInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "project",
		Email:         "member@example.com",
		Role:          domainproject.RoleAdmin,
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if repo.createdInvite.ID != "" {
		t.Fatalf("expected no invite to be created, got %q", repo.createdInvite.ID)
	}
}

func TestAcceptInviteResolvesInviteForCurrentUser(t *testing.T) {
	repo := newProjectServiceRepository()
	service := NewService(repo, nil, &recordingProjectEventRecorder{})

	invite, err := service.AcceptInvite(context.Background(), ResolveInviteInput{
		CurrentUserID: testMemberUserID,
		InviteID:      testInviteID,
	})
	if err != nil {
		t.Fatalf("expected invite accept to succeed: %v", err)
	}
	if invite.Status != domainproject.InviteStatusAccepted {
		t.Fatalf("expected accepted invite, got %q", invite.Status)
	}
	if repo.acceptedInviteID != testInviteID || repo.acceptedUserID != testMemberUserID {
		t.Fatalf("expected accept call for invite %q and user %q, got invite %q user %q", testInviteID, testMemberUserID, repo.acceptedInviteID, repo.acceptedUserID)
	}
}

func TestCancelInviteRequiresManageMembersAndResolvesPendingInvite(t *testing.T) {
	repo := newProjectServiceRepository()
	repo.members[testCurrentUserID] = domainproject.Member{
		ID:        "55555555-5555-5555-5555-555555555555",
		ProjectID: testProjectID,
		UserID:    testCurrentUserID,
		Role:      domainproject.RoleAdmin,
	}
	service := NewService(repo, nil, &recordingProjectEventRecorder{})

	invite, err := service.CancelInvite(context.Background(), CancelInviteInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "project",
		InviteID:      testInviteID,
	})
	if err != nil {
		t.Fatalf("expected invite cancel to succeed: %v", err)
	}
	if invite.Status != domainproject.InviteStatusRejected {
		t.Fatalf("expected rejected invite, got %q", invite.Status)
	}
	if repo.canceledInviteID != testInviteID || repo.canceledProjectID != testProjectID {
		t.Fatalf("expected cancel call for project %q invite %q, got project %q invite %q", testProjectID, testInviteID, repo.canceledProjectID, repo.canceledInviteID)
	}
}

func TestRemoveMemberAllowsNonOwnerSelfLeave(t *testing.T) {
	repo := newProjectServiceRepository()
	repo.members[testCurrentUserID] = domainproject.Member{
		ID:        "55555555-5555-5555-5555-555555555555",
		ProjectID: testProjectID,
		UserID:    testCurrentUserID,
		Role:      domainproject.RoleViewer,
	}
	service := NewService(repo, nil, &recordingProjectEventRecorder{})

	err := service.RemoveMember(context.Background(), RemoveMemberInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "project",
		UserID:        testCurrentUserID,
	})
	if err != nil {
		t.Fatalf("expected self leave to succeed: %v", err)
	}
	if repo.deletedUserID != testCurrentUserID {
		t.Fatalf("expected deleted user %q, got %q", testCurrentUserID, repo.deletedUserID)
	}
	if repo.getMemberRoleCalls != 0 {
		t.Fatalf("expected self leave to skip manage-member role lookup, got %d calls", repo.getMemberRoleCalls)
	}
}

func TestRemoveMemberRejectsOwnerSelfLeave(t *testing.T) {
	repo := newProjectServiceRepository()
	repo.members[testCurrentUserID] = domainproject.Member{
		ID:        "55555555-5555-5555-5555-555555555555",
		ProjectID: testProjectID,
		UserID:    testCurrentUserID,
		Role:      domainproject.RoleOwner,
	}
	repo.members[testOtherOwnerID] = domainproject.Member{
		ID:        "66666666-6666-6666-6666-666666666666",
		ProjectID: testProjectID,
		UserID:    testOtherOwnerID,
		Role:      domainproject.RoleOwner,
	}
	service := NewService(repo, nil, &recordingProjectEventRecorder{})

	err := service.RemoveMember(context.Background(), RemoveMemberInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "project",
		UserID:        testCurrentUserID,
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden owner self leave, got %v", err)
	}
	if repo.deletedUserID != "" {
		t.Fatalf("expected owner self leave not to delete a member, deleted %q", repo.deletedUserID)
	}
	if repo.countOwnersCalls != 0 {
		t.Fatalf("expected owner self leave to fail before owner counting, got %d calls", repo.countOwnersCalls)
	}
}

type projectServiceRepository struct {
	project            domainproject.Project
	members            map[string]domainproject.Member
	createdInvite      domainproject.Invite
	deletedProjectID   string
	deletedUserID      string
	acceptedInviteID   string
	acceptedUserID     string
	rejectedInviteID   string
	rejectedUserID     string
	canceledProjectID  string
	canceledInviteID   string
	getMemberRoleCalls int
	countOwnersCalls   int
}

func newProjectServiceRepository() *projectServiceRepository {
	return &projectServiceRepository{
		project: domainproject.Project{
			ID:              testProjectID,
			Name:            "Project",
			Slug:            "project",
			CreatedByUserID: testCurrentUserID,
		},
		members: map[string]domainproject.Member{},
	}
}

func (r *projectServiceRepository) CreateProjectWithOwner(context.Context, domainproject.Project) (domainproject.Project, error) {
	return domainproject.Project{}, nil
}

func (r *projectServiceRepository) ListProjectsForUser(context.Context, string) ([]domainproject.Project, error) {
	return nil, nil
}

func (r *projectServiceRepository) GetProjectForUser(context.Context, string, string) (domainproject.Project, error) {
	return r.project, nil
}

func (r *projectServiceRepository) GetMemberRole(_ context.Context, _, userID string) (domainproject.Role, error) {
	r.getMemberRoleCalls++
	member, ok := r.members[userID]
	if !ok {
		return "", domainproject.ErrMemberNotFound
	}

	return member.Role, nil
}

func (r *projectServiceRepository) UpdateProject(_ context.Context, input domainproject.Project) (domainproject.Project, error) {
	return input, nil
}

func (r *projectServiceRepository) SoftDeleteProject(_ context.Context, projectID string) error {
	r.deletedProjectID = projectID
	return nil
}

func (r *projectServiceRepository) ListMembers(context.Context, string) ([]domainproject.Member, error) {
	members := make([]domainproject.Member, 0, len(r.members))
	for _, member := range r.members {
		members = append(members, member)
	}

	return members, nil
}

func (r *projectServiceRepository) GetMember(_ context.Context, _, userID string) (domainproject.Member, error) {
	member, ok := r.members[userID]
	if !ok {
		return domainproject.Member{}, domainproject.ErrMemberNotFound
	}

	return member, nil
}

func (r *projectServiceRepository) UpdateMemberRole(_ context.Context, input domainproject.Member) (domainproject.Member, error) {
	return input, nil
}

func (r *projectServiceRepository) DeleteMember(_ context.Context, _, userID string) error {
	r.deletedUserID = userID
	return nil
}

func (r *projectServiceRepository) CountOwners(context.Context, string) (int, error) {
	r.countOwnersCalls++
	owners := 0
	for _, member := range r.members {
		if member.Role == domainproject.RoleOwner {
			owners++
		}
	}

	return owners, nil
}

func (r *projectServiceRepository) CreateInvite(_ context.Context, input domainproject.Invite) (domainproject.Invite, error) {
	invite := r.hydrateInvite(input, domainproject.InviteStatusPending)
	r.createdInvite = invite
	return invite, nil
}

func (r *projectServiceRepository) ListProjectInvites(context.Context, string) ([]domainproject.Invite, error) {
	if r.createdInvite.ID == "" {
		return nil, nil
	}

	return []domainproject.Invite{r.createdInvite}, nil
}

func (r *projectServiceRepository) ListUserInvites(context.Context, string) ([]domainproject.Invite, error) {
	if r.createdInvite.ID == "" {
		return nil, nil
	}

	return []domainproject.Invite{r.createdInvite}, nil
}

func (r *projectServiceRepository) CancelInvite(_ context.Context, projectID, inviteID string) (domainproject.Invite, error) {
	r.canceledProjectID = projectID
	r.canceledInviteID = inviteID

	return r.hydrateInvite(domainproject.Invite{
		ID:              inviteID,
		ProjectID:       projectID,
		InvitedUserID:   testMemberUserID,
		InvitedByUserID: testCurrentUserID,
		Role:            domainproject.RoleViewer,
	}, domainproject.InviteStatusRejected), nil
}

func (r *projectServiceRepository) AcceptInvite(_ context.Context, inviteID, userID string) (domainproject.Invite, error) {
	r.acceptedInviteID = inviteID
	r.acceptedUserID = userID

	return r.hydrateInvite(domainproject.Invite{
		ID:              inviteID,
		ProjectID:       testProjectID,
		InvitedUserID:   userID,
		InvitedByUserID: testCurrentUserID,
		Role:            domainproject.RoleViewer,
	}, domainproject.InviteStatusAccepted), nil
}

func (r *projectServiceRepository) RejectInvite(_ context.Context, inviteID, userID string) (domainproject.Invite, error) {
	r.rejectedInviteID = inviteID
	r.rejectedUserID = userID

	return r.hydrateInvite(domainproject.Invite{
		ID:              inviteID,
		ProjectID:       testProjectID,
		InvitedUserID:   userID,
		InvitedByUserID: testCurrentUserID,
		Role:            domainproject.RoleViewer,
	}, domainproject.InviteStatusRejected), nil
}

func (r *projectServiceRepository) hydrateInvite(input domainproject.Invite, status domainproject.InviteStatus) domainproject.Invite {
	invite := input
	if invite.ID == "" {
		invite.ID = testInviteID
	}
	if invite.ProjectID == "" {
		invite.ProjectID = testProjectID
	}
	if invite.Status == "" {
		invite.Status = status
	}
	if invite.InvitedEmail == "" {
		invite.InvitedEmail = "member@example.com"
	}
	invite.Project = domainproject.InviteProject{
		ID:   invite.ProjectID,
		Name: r.project.Name,
		Slug: r.project.Slug,
	}
	invite.InvitedUser = domainproject.InviteUser{
		ID:          invite.InvitedUserID,
		Email:       invite.InvitedEmail,
		DisplayName: "Member",
	}
	invite.InvitedByUser = domainproject.MemberUser{
		ID:          invite.InvitedByUserID,
		Email:       "owner@example.com",
		DisplayName: "Owner",
	}

	return invite
}

type projectServiceUserLookup struct {
	user identity.User
	err  error
}

func (l *projectServiceUserLookup) GetUserByEmail(context.Context, string) (identity.User, error) {
	if l.err != nil {
		return identity.User{}, l.err
	}

	return l.user, nil
}

type recordingProjectEventRecorder struct {
	events []ProjectEvent
}

func (r *recordingProjectEventRecorder) RecordProjectEvent(_ context.Context, event ProjectEvent) {
	r.events = append(r.events, event)
}
