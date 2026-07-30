package project

import (
	"context"
	"errors"
	"testing"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func TestCreateProjectHonorsProjectCreationPolicy(t *testing.T) {
	tests := map[string]struct {
		policy          projectCreationPolicyStub
		wantErr         error
		wantCreateCalls int
		wantReason      ProjectEventReason
		wantTechnical   bool
	}{
		"enabled": {
			policy:          projectCreationPolicyStub{enabled: true},
			wantCreateCalls: 1,
		},
		"disabled": {
			policy:     projectCreationPolicyStub{enabled: false},
			wantErr:    ErrForbidden,
			wantReason: ProjectReasonForbidden,
		},
		"lookup failure": {
			policy:        projectCreationPolicyStub{err: errProjectPolicyUnavailable},
			wantErr:       errProjectPolicyUnavailable,
			wantReason:    ProjectReasonPolicyLookupFailed,
			wantTechnical: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			repo := &projectPolicyRepository{projectServiceRepository: newProjectServiceRepository()}
			events := &recordingProjectEventRecorder{}
			service := NewService(repo, nil, events)
			service.ConfigureInstancePolicy(&test.policy)

			_, err := service.CreateProject(context.Background(), CreateProjectInput{
				CurrentUserID: testCurrentUserID,
				Name:          "Project",
				Slug:          "project",
			})

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("expected error %v, got %v", test.wantErr, err)
			}
			if test.wantErr != nil && errors.Is(err, ErrForbidden) != errors.Is(test.wantErr, ErrForbidden) {
				t.Fatalf("expected policy lookup failure not to be reported as forbidden")
			}
			if test.policy.calls != 1 {
				t.Fatalf("expected one policy lookup, got %d", test.policy.calls)
			}
			if repo.createCalls != test.wantCreateCalls {
				t.Fatalf("expected %d project creation calls, got %d", test.wantCreateCalls, repo.createCalls)
			}

			if test.wantErr == nil {
				return
			}
			if len(events.events) != 1 {
				t.Fatalf("expected one failure event, got %d", len(events.events))
			}
			event := events.events[0]
			if event.Reason != test.wantReason {
				t.Fatalf("expected failure reason %q, got %q", test.wantReason, event.Reason)
			}
			if test.wantTechnical && !errors.Is(event.Err, test.wantErr) {
				t.Fatalf("expected technical event to retain policy error, got %v", event.Err)
			}
			if !test.wantTechnical && event.Err != nil {
				t.Fatalf("expected business failure not to retain internal error, got %v", event.Err)
			}
		})
	}
}

var errProjectPolicyUnavailable = errors.New("project creation policy unavailable")

type projectCreationPolicyStub struct {
	enabled bool
	err     error
	calls   int
}

func (p *projectCreationPolicyStub) ProjectCreationEnabled(context.Context) (bool, error) {
	p.calls++
	return p.enabled, p.err
}

type projectPolicyRepository struct {
	*projectServiceRepository
	createCalls int
}

func (r *projectPolicyRepository) CreateProjectWithOwner(_ context.Context, input domainproject.Project) (domainproject.Project, error) {
	r.createCalls++
	input.ID = testProjectID
	return input, nil
}
