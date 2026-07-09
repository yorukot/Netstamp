package project

import (
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func normalizeCreateProjectInput(input CreateProjectInput) (CreateProjectInput, error) {
	var validation appvalidation.Collector

	name, err := domainproject.VNProjectName(input.Name)
	if err != nil {
		validation.AddError("name", err, input.Name)
	}
	slug, err := domainproject.VNProjectSlug(input.Slug)
	if err != nil {
		validation.AddError("slug", err, input.Slug)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return CreateProjectInput{}, err
	}

	return CreateProjectInput{
		CurrentUserID: input.CurrentUserID,
		Name:          name,
		Slug:          slug,
	}, nil
}

func normalizeUpdateProjectInput(input UpdateProjectInput) (UpdateProjectInput, error) {
	output := UpdateProjectInput{
		CurrentUserID: input.CurrentUserID,
	}
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	} else {
		output.ProjectRef = projectRef
	}

	if input.Name == nil && input.Slug == nil {
		validation.Add("", "at least one field must be provided", nil)
		return UpdateProjectInput{}, validation.Err(ErrInvalidInput)
	}

	if input.Name != nil {
		name, err := domainproject.VNProjectName(*input.Name)
		if err != nil {
			validation.AddError("name", err, input.Name)
		} else {
			output.Name = &name
		}
	}
	if input.Slug != nil {
		slug, err := domainproject.VNProjectSlug(*input.Slug)
		if err != nil {
			validation.AddError("slug", err, input.Slug)
		} else {
			output.Slug = &slug
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return UpdateProjectInput{}, err
	}

	return output, nil
}

func normalizeCreateInviteInput(input CreateInviteInput) (CreateInviteInput, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	email, err := identity.VNUserEmail(input.Email)
	if err != nil {
		validation.AddError("email", err, input.Email)
	}
	role, err := domainproject.VNProjectMemberRole(input.Role)
	if err != nil {
		validation.AddError("role", err, input.Role)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return CreateInviteInput{}, err
	}

	return CreateInviteInput{
		CurrentUserID: input.CurrentUserID,
		ProjectRef:    projectRef,
		Email:         email,
		Role:          role,
	}, nil
}

func normalizeResolveInviteInput(input ResolveInviteInput) (ResolveInviteInput, error) {
	var validation appvalidation.Collector

	inviteID, err := domainproject.VNProjectInviteID(input.InviteID)
	if err != nil {
		validation.AddError("inviteId", err, input.InviteID)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return ResolveInviteInput{}, err
	}

	return ResolveInviteInput{
		CurrentUserID: input.CurrentUserID,
		InviteID:      inviteID,
	}, nil
}

func normalizeCancelInviteInput(input CancelInviteInput) (CancelInviteInput, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	inviteID, err := domainproject.VNProjectInviteID(input.InviteID)
	if err != nil {
		validation.AddError("inviteId", err, input.InviteID)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return CancelInviteInput{}, err
	}

	return CancelInviteInput{
		CurrentUserID: input.CurrentUserID,
		ProjectRef:    projectRef,
		InviteID:      inviteID,
	}, nil
}

func normalizeUpdateMemberRoleInput(input UpdateMemberRoleInput) (UpdateMemberRoleInput, error) {
	projectRef, userID, role, err := normalizeMemberRoleFields(input.ProjectRef, input.UserID, input.Role, "memberUserId")
	if err != nil {
		return UpdateMemberRoleInput{}, err
	}

	return UpdateMemberRoleInput{
		CurrentUserID: input.CurrentUserID,
		ProjectRef:    projectRef,
		UserID:        userID,
		Role:          role,
	}, nil
}

func normalizeMemberRoleFields(projectRefValue, userIDValue string, roleValue domainproject.Role, userIDField string) (string, string, domainproject.Role, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(projectRefValue)
	if err != nil {
		validation.AddError("projectRef", err, projectRefValue)
	}
	userID, err := domainproject.VNProjectMemberUserID(userIDValue)
	if err != nil {
		validation.AddError(userIDField, err, userIDValue)
	}
	role, err := domainproject.VNProjectMemberRole(roleValue)
	if err != nil {
		validation.AddError("role", err, roleValue)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return "", "", "", err
	}

	return projectRef, userID, role, nil
}

func normalizeRemoveMemberInput(input RemoveMemberInput) (RemoveMemberInput, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	userID, err := domainproject.VNProjectMemberUserID(input.UserID)
	if err != nil {
		validation.AddError("memberUserId", err, input.UserID)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return RemoveMemberInput{}, err
	}

	return RemoveMemberInput{
		CurrentUserID: input.CurrentUserID,
		ProjectRef:    projectRef,
		UserID:        userID,
	}, nil
}

func invalidProjectField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
