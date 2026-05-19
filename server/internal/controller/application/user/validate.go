package account

import (
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func normalizeUpdateCurrentUserInput(input UpdateCurrentUserInput) (UpdateCurrentUserInput, error) {
	output := UpdateCurrentUserInput{}
	var validation appvalidation.Collector

	userID, err := identity.VNUserID(input.CurrentUserID)
	if err != nil {
		validation.AddError("currentUserId", err, input.CurrentUserID)
	} else {
		output.CurrentUserID = userID
	}

	if input.DisplayName == nil {
		validation.Add("", "at least one field must be provided", nil)
		return UpdateCurrentUserInput{}, validation.Err(ErrInvalidInput)
	}

	displayName, err := identity.VNUserDisplayName(*input.DisplayName)
	if err != nil {
		validation.AddError("displayName", err, input.DisplayName)
	} else {
		output.DisplayName = &displayName
	}

	if err := validation.Err(ErrInvalidInput); err != nil {
		return UpdateCurrentUserInput{}, err
	}

	return output, nil
}

func normalizeChangeCurrentUserEmailInput(input ChangeCurrentUserEmailInput) (ChangeCurrentUserEmailInput, error) {
	var validation appvalidation.Collector

	userID, err := identity.VNUserID(input.CurrentUserID)
	if err != nil {
		validation.AddError("currentUserId", err, input.CurrentUserID)
	}
	newEmail, err := identity.VNUserEmail(input.NewEmail)
	if err != nil {
		validation.AddError("newEmail", err, input.NewEmail)
	}
	password, err := identity.VNUserPassword(input.Password)
	if err != nil {
		validation.AddError("password", err, "")
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return ChangeCurrentUserEmailInput{}, err
	}

	return ChangeCurrentUserEmailInput{
		CurrentUserID: userID,
		NewEmail:      newEmail,
		Password:      password,
	}, nil
}

func normalizeChangeCurrentUserPasswordInput(input ChangeCurrentUserPasswordInput) (ChangeCurrentUserPasswordInput, error) {
	var validation appvalidation.Collector

	userID, err := identity.VNUserID(input.CurrentUserID)
	if err != nil {
		validation.AddError("currentUserId", err, input.CurrentUserID)
	}
	currentPassword, err := identity.VNUserPassword(input.CurrentPassword)
	if err != nil {
		validation.AddError("currentPassword", err, "")
	}
	newPassword, err := identity.VNUserPassword(input.NewPassword)
	if err != nil {
		validation.AddError("newPassword", err, "")
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return ChangeCurrentUserPasswordInput{}, err
	}

	return ChangeCurrentUserPasswordInput{
		CurrentUserID:   userID,
		CurrentPassword: currentPassword,
		NewPassword:     newPassword,
	}, nil
}
