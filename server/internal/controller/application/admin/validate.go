package admin

import (
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainsystem "github.com/yorukot/netstamp/internal/domain/system"
)

const (
	auditActionGrantSystemAdmin  = "grant_system_admin"
	auditActionRevokeSystemAdmin = "revoke_system_admin"
	auditActionDisableUser       = "disable_user"
	auditActionEnableUser        = "enable_user"
	auditActionSetPassword       = "set_user_password"
	auditActionClearPassword     = "clear_user_password"
)

func normalizeGrantSystemAdminInput(input GrantSystemAdminInput) (GrantSystemAdminInput, error) {
	currentUserID, currentUserErr := identity.VNUserID(input.CurrentUserID)
	email, emailErr := identity.VNUserEmail(input.Email)
	if currentUserErr == nil && emailErr == nil {
		return GrantSystemAdminInput{CurrentUserID: currentUserID, Email: email}, nil
	}

	collector := appvalidation.Collector{}
	collector.AddError("currentUserId", currentUserErr, input.CurrentUserID)
	collector.AddError("email", emailErr, input.Email)
	return GrantSystemAdminInput{}, collector.Err(ErrInvalidInput)
}

func normalizeRevokeSystemAdminInput(input RevokeSystemAdminInput) (RevokeSystemAdminInput, error) {
	currentUserID, currentUserErr := identity.VNUserID(input.CurrentUserID)
	userID, userErr := identity.VNUserID(input.UserID)
	if currentUserErr == nil && userErr == nil {
		return RevokeSystemAdminInput{CurrentUserID: currentUserID, UserID: userID}, nil
	}

	collector := appvalidation.Collector{}
	collector.AddError("currentUserId", currentUserErr, input.CurrentUserID)
	collector.AddError("userId", userErr, input.UserID)
	return RevokeSystemAdminInput{}, collector.Err(ErrInvalidInput)
}

func normalizeListManagedUsersInput(input ListManagedUsersInput) (ListManagedUsersInput, error) {
	currentUserID, currentUserErr := identity.VNUserID(input.CurrentUserID)
	if currentUserErr == nil {
		return ListManagedUsersInput{CurrentUserID: currentUserID}, nil
	}

	collector := appvalidation.Collector{}
	collector.AddError("currentUserId", currentUserErr, input.CurrentUserID)
	return ListManagedUsersInput{}, collector.Err(ErrInvalidInput)
}

func normalizeUpdateManagedUserInput(input UpdateManagedUserInput) (UpdateManagedUserInput, error) {
	currentUserID, currentUserErr := identity.VNUserID(input.CurrentUserID)
	userID, userErr := identity.VNUserID(input.UserID)
	if currentUserErr == nil && userErr == nil {
		input.CurrentUserID = currentUserID
		input.UserID = userID
		return input, nil
	}

	collector := appvalidation.Collector{}
	collector.AddError("currentUserId", currentUserErr, input.CurrentUserID)
	collector.AddError("userId", userErr, input.UserID)
	return UpdateManagedUserInput{}, collector.Err(ErrInvalidInput)
}

func normalizeSetManagedUserPasswordInput(input SetManagedUserPasswordInput) (SetManagedUserPasswordInput, error) {
	currentUserID, currentUserErr := identity.VNUserID(input.CurrentUserID)
	userID, userErr := identity.VNUserID(input.UserID)
	password, passwordErr := identity.VNUserPassword(input.Password)
	if currentUserErr == nil && userErr == nil && passwordErr == nil {
		return SetManagedUserPasswordInput{
			CurrentUserID: currentUserID,
			UserID:        userID,
			Password:      password,
		}, nil
	}

	collector := appvalidation.Collector{}
	collector.AddError("currentUserId", currentUserErr, input.CurrentUserID)
	collector.AddError("userId", userErr, input.UserID)
	collector.AddError("password", passwordErr, input.Password)
	return SetManagedUserPasswordInput{}, collector.Err(ErrInvalidInput)
}

func normalizeClearManagedUserPasswordInput(input ClearManagedUserPasswordInput) (ClearManagedUserPasswordInput, error) {
	currentUserID, currentUserErr := identity.VNUserID(input.CurrentUserID)
	userID, userErr := identity.VNUserID(input.UserID)
	if currentUserErr == nil && userErr == nil {
		return ClearManagedUserPasswordInput{CurrentUserID: currentUserID, UserID: userID}, nil
	}

	collector := appvalidation.Collector{}
	collector.AddError("currentUserId", currentUserErr, input.CurrentUserID)
	collector.AddError("userId", userErr, input.UserID)
	return ClearManagedUserPasswordInput{}, collector.Err(ErrInvalidInput)
}

func systemAdminAuditKey(admin domainsystem.AdminUser) string {
	return "system_admin:" + admin.ID
}

func managedUserAuditKey(user ManagedUser) string {
	return "user:" + user.ID
}
