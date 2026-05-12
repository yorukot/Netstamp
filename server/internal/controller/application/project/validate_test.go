package project

import (
	"testing"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

const (
	testCurrentUserID = "11111111-1111-1111-1111-111111111111"
	testMemberUserID  = "22222222-2222-2222-2222-222222222222"
)

func TestNormalizeCreateProjectInputPreservesCurrentUserID(t *testing.T) {
	input, err := normalizeCreateProjectInput(CreateProjectInput{
		CurrentUserID: testCurrentUserID,
		Name:          " Project ",
		Slug:          "project",
	})
	if err != nil {
		t.Fatalf("expected valid input: %v", err)
	}

	if input.CurrentUserID != testCurrentUserID {
		t.Fatalf("expected current user ID to be preserved, got %q", input.CurrentUserID)
	}
}

func TestNormalizeUpdateProjectInputPreservesContext(t *testing.T) {
	name := " Project "
	input, err := normalizeUpdateProjectInput(UpdateProjectInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "project",
		Name:          &name,
	})
	if err != nil {
		t.Fatalf("expected valid input: %v", err)
	}

	if input.CurrentUserID != testCurrentUserID {
		t.Fatalf("expected current user ID to be preserved, got %q", input.CurrentUserID)
	}
	if input.ProjectRef != "project" {
		t.Fatalf("expected project ref to be preserved, got %q", input.ProjectRef)
	}
}

func TestNormalizeMemberInputsPreserveCurrentUserID(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		input, err := normalizeAddMemberInput(AddMemberInput{
			CurrentUserID: testCurrentUserID,
			ProjectRef:    " project ",
			UserID:        testMemberUserID,
			Role:          domainproject.RoleEditor,
		})
		if err != nil {
			t.Fatalf("expected valid input: %v", err)
		}
		if input.CurrentUserID != testCurrentUserID {
			t.Fatalf("expected current user ID to be preserved, got %q", input.CurrentUserID)
		}
	})

	t.Run("update role", func(t *testing.T) {
		input, err := normalizeUpdateMemberRoleInput(UpdateMemberRoleInput{
			CurrentUserID: testCurrentUserID,
			ProjectRef:    " project ",
			UserID:        testMemberUserID,
			Role:          domainproject.RoleViewer,
		})
		if err != nil {
			t.Fatalf("expected valid input: %v", err)
		}
		if input.CurrentUserID != testCurrentUserID {
			t.Fatalf("expected current user ID to be preserved, got %q", input.CurrentUserID)
		}
	})

	t.Run("remove", func(t *testing.T) {
		input, err := normalizeRemoveMemberInput(RemoveMemberInput{
			CurrentUserID: testCurrentUserID,
			ProjectRef:    " project ",
			UserID:        testMemberUserID,
		})
		if err != nil {
			t.Fatalf("expected valid input: %v", err)
		}
		if input.CurrentUserID != testCurrentUserID {
			t.Fatalf("expected current user ID to be preserved, got %q", input.CurrentUserID)
		}
	})
}
