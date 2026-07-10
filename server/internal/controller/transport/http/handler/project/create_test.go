package project

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

const projectHandlerTestUserID = "11111111-1111-1111-1111-111111111111"

func TestCreateProjectReturnsConflictProblemForSlugConflict(t *testing.T) {
	router := chi.NewRouter()
	service := appproject.NewService(
		&projectHandlerRepository{createErr: domainproject.ErrProjectSlugAlreadyExists},
		nil,
		&projectHandlerEventRecorder{},
	)
	NewHandler(service, &projectHandlerTokenVerifier{
		claims: identity.SessionClaims{
			SessionID: "session-1",
			UserID:    projectHandlerTestUserID,
		},
	}, "netstamp_session").RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(`{"name":"Project","slug":"project"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "netstamp_session=valid-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", res.Code)
	}
	if got := res.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("expected problem content type, got %q", got)
	}

	var body httpx.ProblemDetails
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != http.StatusConflict {
		t.Fatalf("expected problem status 409, got %d", body.Status)
	}
	if body.Code != httpx.CodeProjectSlugAlreadyExists {
		t.Fatalf("expected slug conflict code, got %q", body.Code)
	}
	if body.Detail != "project slug already exists" {
		t.Fatalf("expected slug conflict detail, got %q", body.Detail)
	}
}

func TestMapProjectErrorUsesSpecificNotFoundDetails(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		detail string
		code   string
	}{
		{name: "project", err: domainproject.ErrProjectNotFound, detail: "project not found", code: httpx.CodeProjectNotFound},
		{name: "member", err: domainproject.ErrMemberNotFound, detail: "project member not found", code: httpx.CodeProjectMemberNotFound},
		{name: "invite", err: domainproject.ErrInviteNotFound, detail: "project invite not found", code: httpx.CodeProjectInviteNotFound},
		{name: "user", err: identity.ErrUserNotFound, detail: "user not found", code: httpx.CodeUserNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := mapProjectError(test.err, "fallback")

			var httpErr *httpx.Error
			if !errors.As(err, &httpErr) {
				t.Fatalf("expected http error, got %T", err)
			}
			if httpErr.Status != http.StatusNotFound {
				t.Fatalf("expected status 404, got %d", httpErr.Status)
			}
			if httpErr.Detail != test.detail {
				t.Fatalf("expected detail %q, got %q", test.detail, httpErr.Detail)
			}
			if httpErr.Code != test.code {
				t.Fatalf("expected code %q, got %q", test.code, httpErr.Code)
			}
		})
	}
}

type projectHandlerTokenVerifier struct {
	claims identity.SessionClaims
}

func (v *projectHandlerTokenVerifier) VerifySession(context.Context, string) (identity.SessionClaims, error) {
	return v.claims, nil
}

func (v *projectHandlerTokenVerifier) CreateSession(context.Context, appauth.CreateSessionInput) (identity.CreatedSession, error) {
	return identity.CreatedSession{}, nil
}

func (v *projectHandlerTokenVerifier) CreateCSRFToken(context.Context, string) (string, error) {
	return "", nil
}

func (v *projectHandlerTokenVerifier) VerifyCSRFToken(context.Context, string, string) error {
	return nil
}

func (v *projectHandlerTokenVerifier) RevokeSession(context.Context, string, string) error {
	return nil
}

func (v *projectHandlerTokenVerifier) RevokeUserSessions(context.Context, string, string) error {
	return nil
}

var _ appauth.SessionManager = (*projectHandlerTokenVerifier)(nil)

type projectHandlerEventRecorder struct{}

func (r *projectHandlerEventRecorder) RecordProjectEvent(context.Context, appproject.ProjectEvent) {}

type projectHandlerRepository struct {
	createErr error
}

func (r *projectHandlerRepository) CreateProjectWithOwner(context.Context, domainproject.Project) (domainproject.Project, error) {
	if r.createErr != nil {
		return domainproject.Project{}, r.createErr
	}

	return domainproject.Project{
		ID:              "22222222-2222-2222-2222-222222222222",
		Name:            "Project",
		Slug:            "project",
		CreatedByUserID: projectHandlerTestUserID,
	}, nil
}

func (r *projectHandlerRepository) ListProjectsForUser(context.Context, string) ([]domainproject.Project, error) {
	return nil, nil
}

func (r *projectHandlerRepository) GetProjectForUser(context.Context, string, string) (domainproject.Project, error) {
	return domainproject.Project{}, nil
}

func (r *projectHandlerRepository) GetMemberRole(context.Context, string, string) (domainproject.Role, error) {
	return "", nil
}

func (r *projectHandlerRepository) UpdateProject(context.Context, domainproject.Project) (domainproject.Project, error) {
	return domainproject.Project{}, nil
}

func (r *projectHandlerRepository) SoftDeleteProject(context.Context, string) error {
	return nil
}

func (r *projectHandlerRepository) ListMembers(context.Context, string) ([]domainproject.Member, error) {
	return nil, nil
}

func (r *projectHandlerRepository) GetMember(context.Context, string, string) (domainproject.Member, error) {
	return domainproject.Member{}, nil
}

func (r *projectHandlerRepository) UpdateMemberRole(context.Context, domainproject.Member) (domainproject.Member, error) {
	return domainproject.Member{}, nil
}

func (r *projectHandlerRepository) DeleteMember(context.Context, string, string) error {
	return nil
}

func (r *projectHandlerRepository) CountOwners(context.Context, string) (int, error) {
	return 0, nil
}

func (r *projectHandlerRepository) CreateInvite(context.Context, domainproject.Invite) (domainproject.Invite, error) {
	return domainproject.Invite{}, nil
}

func (r *projectHandlerRepository) ListProjectInvites(context.Context, string) ([]domainproject.Invite, error) {
	return nil, nil
}

func (r *projectHandlerRepository) ListUserInvites(context.Context, string) ([]domainproject.Invite, error) {
	return nil, nil
}

func (r *projectHandlerRepository) CancelInvite(context.Context, string, string) (domainproject.Invite, error) {
	return domainproject.Invite{}, nil
}

func (r *projectHandlerRepository) AcceptInvite(context.Context, string, string) (domainproject.Invite, error) {
	return domainproject.Invite{}, nil
}

func (r *projectHandlerRepository) RejectInvite(context.Context, string, string) (domainproject.Invite, error) {
	return domainproject.Invite{}, nil
}
