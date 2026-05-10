package project

import (
	appvalidation "github.com/yorukot/netstamp/internal/application/validation"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"github.com/yorukot/netstamp/internal/normalize"
)

const (
	maxProjectNameRunes = 100
	maxProjectSlugRunes = 100
	maxProjectRefRunes  = 100
)

type normalizedCreateProjectInput struct {
	name string
	slug string
}

type normalizedUpdateProjectInput struct {
	name *string
	slug *string
}

type normalizedMemberInput struct {
	projectRef string
	userID     string
	role       domainproject.Role
}

func normalizeCreateProjectInput(input CreateProjectInput) (normalizedCreateProjectInput, error) {
	name, err := appvalidation.RequiredString(ErrInvalidInput, "name", input.Name, maxProjectNameRunes)
	if err != nil {
		return normalizedCreateProjectInput{}, err
	}
	slug, err := normalizeProjectSlug(input.Slug)
	if err != nil {
		return normalizedCreateProjectInput{}, err
	}

	return normalizedCreateProjectInput{name: name, slug: slug}, nil
}

func normalizeUpdateProjectInput(input UpdateProjectInput) (normalizedUpdateProjectInput, error) {
	if input.Name == nil && input.Slug == nil {
		return normalizedUpdateProjectInput{}, invalidProjectField("", "at least one field must be provided", nil)
	}

	name, err := appvalidation.OptionalString(ErrInvalidInput, "name", input.Name, maxProjectNameRunes)
	if err != nil {
		return normalizedUpdateProjectInput{}, err
	}
	slug, err := normalizeOptionalProjectSlug(input.Slug)
	if err != nil {
		return normalizedUpdateProjectInput{}, err
	}

	return normalizedUpdateProjectInput{name: name, slug: slug}, nil
}

func normalizeAddMemberInput(input AddMemberInput) (normalizedMemberInput, error) {
	projectRef, err := normalizeProjectRef(input.ProjectRef)
	if err != nil {
		return normalizedMemberInput{}, err
	}
	userID, err := appvalidation.CanonicalUUID(ErrInvalidInput, "userId", input.UserID)
	if err != nil {
		return normalizedMemberInput{}, err
	}
	role, err := normalizeProjectRole(input.Role)
	if err != nil {
		return normalizedMemberInput{}, err
	}

	return normalizedMemberInput{projectRef: projectRef, userID: userID, role: role}, nil
}

func normalizeUpdateMemberRoleInput(input UpdateMemberRoleInput) (normalizedMemberInput, error) {
	projectRef, err := normalizeProjectRef(input.ProjectRef)
	if err != nil {
		return normalizedMemberInput{}, err
	}
	userID, err := appvalidation.CanonicalUUID(ErrInvalidInput, "memberUserId", input.UserID)
	if err != nil {
		return normalizedMemberInput{}, err
	}
	role, err := normalizeProjectRole(input.Role)
	if err != nil {
		return normalizedMemberInput{}, err
	}

	return normalizedMemberInput{projectRef: projectRef, userID: userID, role: role}, nil
}

func normalizeRemoveMemberInput(input RemoveMemberInput) (normalizedMemberInput, error) {
	projectRef, err := normalizeProjectRef(input.ProjectRef)
	if err != nil {
		return normalizedMemberInput{}, err
	}
	userID, err := appvalidation.CanonicalUUID(ErrInvalidInput, "memberUserId", input.UserID)
	if err != nil {
		return normalizedMemberInput{}, err
	}

	return normalizedMemberInput{projectRef: projectRef, userID: userID}, nil
}

func normalizeProjectRef(value string) (string, error) {
	return appvalidation.RequiredString(ErrInvalidInput, "projectRef", value, maxProjectRefRunes)
}

func normalizeProjectSlug(value string) (string, error) {
	normalized, err := appvalidation.RequiredString(ErrInvalidInput, "slug", value, maxProjectSlugRunes)
	if err != nil {
		return "", err
	}
	if slug, err := normalize.ProjectSlug(normalized, ErrInvalidInput); err == nil {
		return slug, nil
	}

	return "", invalidProjectField("slug", "must contain only lowercase letters, numbers, and dashes", value)
}

func normalizeOptionalProjectSlug(value *string) (*string, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // A nil pointer is the explicit representation of an omitted optional field.
	}

	slug, err := normalizeProjectSlug(*value)
	if err != nil {
		return nil, err
	}

	return &slug, nil
}

func normalizeProjectRole(role domainproject.Role) (domainproject.Role, error) {
	if domainproject.IsValidRole(role) {
		return role, nil
	}

	return "", appvalidation.New(ErrInvalidRole, "role", `must be "owner", "admin", "editor", or "viewer"`, role)
}

func invalidProjectField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
