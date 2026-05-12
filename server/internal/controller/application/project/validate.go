package project

import (
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func normalizeCreateProjectInput(input CreateProjectInput) (CreateProjectInput, error) {
	name, err := domainproject.VNProjectName(input.Name)
	if err != nil {
		return CreateProjectInput{}, invalidProjectField("name", err.Error(), input.Name)
	}
	slug, err := domainproject.VNProjectSlug(input.Slug)
	if err != nil {
		return CreateProjectInput{}, invalidProjectField("slug", err.Error(), input.Slug)
	}

	return CreateProjectInput{Name: name, Slug: slug}, nil
}

func normalizeUpdateProjectInput(input UpdateProjectInput) (UpdateProjectInput, error) {
	if input.Name == nil && input.Slug == nil {
		return UpdateProjectInput{}, invalidProjectField("", "at least one field must be provided", nil)
	}
	var output UpdateProjectInput

	if input.Name != nil {
		name, err := domainproject.VNProjectName(*input.Name)
		if err != nil {
			return UpdateProjectInput{}, invalidProjectField("name", err.Error(), input.Name)
		}
		output.Name = &name
	}
	if input.Slug != nil {
		slug, err := domainproject.VNProjectSlug(*input.Slug)
		if err != nil && input.Slug != nil {
			return UpdateProjectInput{}, invalidProjectField("slug", err.Error(), input.Slug)
		}
		output.Slug = &slug
	}

	return output, nil
}

func normalizeAddMemberInput(input AddMemberInput) (AddMemberInput, error) {
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		return AddMemberInput{}, invalidProjectField("projectRef", err.Error(), input.ProjectRef)
	}
	userID, err := domainproject.VNProjectMemberUserID(input.UserID)
	if err != nil {
		return AddMemberInput{}, invalidProjectField("userId", err.Error(), input.UserID)
	}
	role, err := domainproject.VNProjectMemberRole(input.Role)
	if err != nil {
		return AddMemberInput{}, invalidProjectField("role", err.Error(), input.Role)
	}
	return AddMemberInput{ProjectRef: projectRef, UserID: userID, Role: role}, nil
}

func normalizeUpdateMemberRoleInput(input UpdateMemberRoleInput) (UpdateMemberRoleInput, error) {
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		return UpdateMemberRoleInput{}, invalidProjectField("projectRef", err.Error(), input.ProjectRef)
	}
	userID, err := domainproject.VNProjectMemberUserID(input.UserID)
	if err != nil {
		return UpdateMemberRoleInput{}, invalidProjectField("userId", err.Error(), input.UserID)
	}
	role, err := domainproject.VNProjectMemberRole(input.Role)
	if err != nil {
		return UpdateMemberRoleInput{}, invalidProjectField("role", err.Error(), input.Role)
	}

	return UpdateMemberRoleInput{ProjectRef: projectRef, UserID: userID, Role: role}, nil
}

func normalizeRemoveMemberInput(input RemoveMemberInput) (RemoveMemberInput, error) {
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		return RemoveMemberInput{}, invalidProjectField("projectRef", err.Error(), input.ProjectRef)
	}
	userID, err := domainproject.VNProjectMemberUserID(input.UserID)
	if err != nil {
		return RemoveMemberInput{}, invalidProjectField("userId", err.Error(), input.UserID)
	}

	return RemoveMemberInput{ProjectRef: projectRef, UserID: userID}, nil
}

func invalidProjectField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
