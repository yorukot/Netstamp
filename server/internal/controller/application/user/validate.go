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
	if err := validation.Err(ErrInvalidInput); err != nil {
		return ChangeCurrentUserEmailInput{}, err
	}

	return ChangeCurrentUserEmailInput{
		CurrentUserID: userID,
		NewEmail:      newEmail,
	}, nil
}

func normalizeChangeCurrentUserPasswordInput(input ChangeCurrentUserPasswordInput) (ChangeCurrentUserPasswordInput, error) {
	var validation appvalidation.Collector

	userID, err := identity.VNUserID(input.CurrentUserID)
	if err != nil {
		validation.AddError("currentUserId", err, input.CurrentUserID)
	}
	sessionID, err := identity.VNUserID(input.CurrentSessionID)
	if err != nil {
		validation.AddError("currentSessionId", err, input.CurrentSessionID)
	}
	newPassword, err := identity.VNUserPassword(input.NewPassword)
	if err != nil {
		validation.AddError("newPassword", err, "")
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return ChangeCurrentUserPasswordInput{}, err
	}

	return ChangeCurrentUserPasswordInput{
		CurrentUserID:    userID,
		CurrentSessionID: sessionID,
		NewPassword:      newPassword,
	}, nil
}

func normalizeDeactivateCurrentUserInput(input DeactivateCurrentUserInput) (DeactivateCurrentUserInput, error) {
	userID, err := identity.VNUserID(input.CurrentUserID)
	if err == nil {
		return DeactivateCurrentUserInput{CurrentUserID: userID}, nil
	}

	var validation appvalidation.Collector
	validation.AddError("currentUserId", err, input.CurrentUserID)
	return DeactivateCurrentUserInput{}, validation.Err(ErrInvalidInput)
}
