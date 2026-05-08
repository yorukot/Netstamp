package project

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humatest"

	appproject "github.com/yorukot/netstamp/internal/application/project"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

const (
	testUserID    = "11111111-1111-1111-1111-111111111111"
	testProjectID = "22222222-2222-2222-2222-222222222222"
	testMemberID  = "33333333-3333-3333-3333-333333333333"
)

func TestCreateProjectReturnsCreatedProject(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerProjectRepository{}
	NewHandler(appproject.NewService(repo, handlerProjectEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects", map[string]any{
		"name": " Engineering ",
		"slug": "engineering",
	}, "Authorization: Bearer valid-token")

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", res.Code)
	}

	var body projectOutputBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Project.ID != testProjectID {
		t.Fatalf("expected project id, got %q", body.Project.ID)
	}
	if body.Project.Slug != "engineering" {
		t.Fatalf("expected slug, got %q", body.Project.Slug)
	}
	if repo.gotCreateInput.CreatedByUserID != testUserID {
		t.Fatalf("expected current user id, got %q", repo.gotCreateInput.CreatedByUserID)
	}
}

func TestCreateProjectRejectsInvalidSlugPattern(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appproject.NewService(&handlerProjectRepository{}, handlerProjectEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects", map[string]any{
		"name": "Engineering",
		"slug": "Engineering_Project",
	}, "Authorization: Bearer valid-token")

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", res.Code)
	}
}

func TestCreateProjectRequiresBearerToken(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appproject.NewService(&handlerProjectRepository{}, handlerProjectEventRecorder{}), &handlerTokenVerifier{}).RegisterRoutes(api)

	res := api.Post("/projects", map[string]any{
		"name": "Engineering",
		"slug": "engineering",
	})

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
}

func TestGetProjectAcceptsSlugRef(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerProjectRepository{}
	NewHandler(appproject.NewService(repo, handlerProjectEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Get("/projects/engineering", "Authorization: Bearer valid-token")

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	if repo.gotProjectRef != "engineering" {
		t.Fatalf("expected slug ref, got %q", repo.gotProjectRef)
	}
}

func TestDeleteProjectMapsNonOwnerToForbidden(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appproject.NewService(&handlerProjectRepository{
		actorRole: domainproject.RoleAdmin,
	}, handlerProjectEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Delete("/projects/engineering", "Authorization: Bearer valid-token")

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", res.Code)
	}
}

func TestAddMemberRejectsOwnerRoleAsForbidden(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(appproject.NewService(&handlerProjectRepository{
		actorRole: domainproject.RoleOwner,
	}, handlerProjectEventRecorder{}), &handlerTokenVerifier{
		claims: identity.AccessTokenClaims{Subject: testUserID, Email: "user@example.com"},
	}).RegisterRoutes(api)

	res := api.Post("/projects/engineering/members", map[string]any{
		"userId": testMemberID,
		"role":   "owner",
	}, "Authorization: Bearer valid-token")

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", res.Code)
	}
}

type handlerTokenVerifier struct {
	claims identity.AccessTokenClaims
	err    error
}

type handlerProjectEventRecorder struct{}

func (handlerProjectEventRecorder) RecordProjectEvent(context.Context, appproject.ProjectEvent) {}

func (v *handlerTokenVerifier) VerifyAccessToken(context.Context, string) (identity.AccessTokenClaims, error) {
	if v.err != nil {
		return identity.AccessTokenClaims{}, v.err
	}
	return v.claims, nil
}

type handlerProjectRepository struct {
	gotCreateInput         domainproject.CreateProjectStorageInput
	gotProjectRef          string
	actorRole              domainproject.Role
	gotSoftDeleteProjectID string
}

func (r *handlerProjectRepository) CreateProjectWithOwner(_ context.Context, input domainproject.CreateProjectStorageInput) (domainproject.Project, error) {
	r.gotCreateInput = input
	return domainproject.Project{
		ID:              testProjectID,
		Name:            input.Name,
		Slug:            input.Slug,
		CreatedByUserID: input.CreatedByUserID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

func (r *handlerProjectRepository) ListProjectsForUser(context.Context, string) ([]domainproject.Project, error) {
	return []domainproject.Project{{
		ID:              testProjectID,
		Name:            "Engineering",
		Slug:            "engineering",
		CreatedByUserID: testUserID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}}, nil
}

func (r *handlerProjectRepository) GetProjectForUser(_ context.Context, projectRef string, _ string) (domainproject.Project, error) {
	r.gotProjectRef = projectRef
	return domainproject.Project{
		ID:              testProjectID,
		Name:            "Engineering",
		Slug:            "engineering",
		CreatedByUserID: testUserID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

func (r *handlerProjectRepository) GetMemberRole(context.Context, string, string) (domainproject.Role, error) {
	if r.actorRole != "" {
		return r.actorRole, nil
	}
	return domainproject.RoleOwner, nil
}

func (r *handlerProjectRepository) UpdateProject(_ context.Context, input domainproject.UpdateProjectStorageInput) (domainproject.Project, error) {
	return domainproject.Project{
		ID:              input.ProjectID,
		Name:            input.Name,
		Slug:            input.Slug,
		CreatedByUserID: testUserID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

func (r *handlerProjectRepository) SoftDeleteProject(_ context.Context, projectID string) error {
	r.gotSoftDeleteProjectID = projectID
	return nil
}

func (r *handlerProjectRepository) ListMembers(context.Context, string) ([]domainproject.Member, error) {
	return []domainproject.Member{{ID: testMemberID, ProjectID: testProjectID, UserID: testUserID, Email: "user@example.com", Role: domainproject.RoleOwner}}, nil
}

func (r *handlerProjectRepository) GetMember(context.Context, string, string) (domainproject.Member, error) {
	return domainproject.Member{ID: testMemberID, ProjectID: testProjectID, UserID: testMemberID, Email: "member@example.com", Role: domainproject.RoleViewer}, nil
}

func (r *handlerProjectRepository) AddMember(_ context.Context, input domainproject.AddMemberStorageInput) (domainproject.Member, error) {
	return domainproject.Member{ID: testMemberID, ProjectID: input.ProjectID, UserID: input.UserID, Email: "member@example.com", Role: input.Role}, nil
}

func (r *handlerProjectRepository) UpdateMemberRole(_ context.Context, input domainproject.UpdateMemberRoleStorageInput) (domainproject.Member, error) {
	return domainproject.Member{ID: testMemberID, ProjectID: input.ProjectID, UserID: input.UserID, Email: "member@example.com", Role: input.Role}, nil
}

func (r *handlerProjectRepository) CountOwners(context.Context, string) (int, error) {
	return 1, nil
}
