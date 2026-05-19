package project

import (
	"errors"
	"testing"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
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

func TestNormalizeCreateProjectInputReturnsAllFieldErrors(t *testing.T) {
	_, err := normalizeCreateProjectInput(CreateProjectInput{
		CurrentUserID: testCurrentUserID,
		Name:          "",
		Slug:          "",
	})

	assertProjectValidationFields(t, err, []string{"name", "slug"})
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

func TestNormalizeUpdateProjectInputReturnsAllFieldErrors(t *testing.T) {
	name := ""
	slug := ""
	_, err := normalizeUpdateProjectInput(UpdateProjectInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "",
		Name:          &name,
		Slug:          &slug,
	})

	assertProjectValidationFields(t, err, []string{"projectRef", "name", "slug"})
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

func TestNormalizeAddMemberInputReturnsAllFieldErrors(t *testing.T) {
	_, err := normalizeAddMemberInput(AddMemberInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "",
		UserID:        "",
		Role:          "invalid",
	})

	assertProjectValidationFields(t, err, []string{"projectRef", "userId", "role"})
}

func TestNormalizeUpdateMemberRoleInputReturnsAllFieldErrors(t *testing.T) {
	_, err := normalizeUpdateMemberRoleInput(UpdateMemberRoleInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "",
		UserID:        "",
		Role:          "invalid",
	})

	assertProjectValidationFields(t, err, []string{"projectRef", "memberUserId", "role"})
}

func assertProjectValidationFields(t *testing.T, err error, wantFields []string) {
	t.Helper()

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}

	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		t.Fatalf("expected field validation errors, got %v", err)
	}
	if len(fieldErrors) != len(wantFields) {
		t.Fatalf("expected %d field errors, got %d: %#v", len(wantFields), len(fieldErrors), fieldErrors)
	}

	for i, wantField := range wantFields {
		if fieldErrors[i].Field != wantField {
			t.Fatalf("expected field error %d to target %q, got %q", i, wantField, fieldErrors[i].Field)
		}
		if fieldErrors[i].Message == "" {
			t.Fatalf("expected field error %d to include a message", i)
		}
	}
}
