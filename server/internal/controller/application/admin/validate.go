package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainsystem "github.com/yorukot/netstamp/internal/domain/system"
)

const (
	keyRegistrationEnabled       = "auth.registration_enabled"
	keyEmailVerificationRequired = "auth.email_verification_required"
	keyBackendBaseURL            = "http.backend_base_url"
	keyPublicWebBaseURL          = "http.public_web_base_url"
	keySMTPHost                  = "smtp.host"
	keySMTPPort                  = "smtp.port"
	keySMTPUsername              = "smtp.username"
	keySMTPPassword              = "smtp.password"
	keySMTPFrom                  = "smtp.from"
	keySMTPTLSMode               = "smtp.tls_mode"
	keySMTPTimeoutSeconds        = "smtp.timeout_seconds"

	auditActionGrantSystemAdmin  = "grant_system_admin"
	auditActionRevokeSystemAdmin = "revoke_system_admin"
	auditActionDisableUser       = "disable_user"
	auditActionEnableUser        = "enable_user"
	auditActionSetPassword       = "set_user_password"
	auditActionDataImport        = "data_import"
	auditActionUpdate            = "update"
)

func validateSettings(settings Settings) error {
	var errs []error
	if err := validateOptionalHTTPOrigin("backendBaseUrl", settings.BackendBaseURL); err != nil {
		errs = append(errs, err)
	}
	if err := validateOptionalHTTPOrigin("publicWebBaseUrl", settings.PublicWebBaseURL); err != nil {
		errs = append(errs, err)
	}
	if err := validateSMTP(settings.SMTP); err != nil {
		errs = append(errs, err)
	}
	if settings.EmailVerificationRequired && !smtpDeliveryConfigured(settings.SMTP) {
		errs = append(errs, errors.New("emailVerificationRequired requires smtp.host and smtp.from"))
	}
	if len(errs) > 0 {
		return errors.Join(append([]error{ErrInvalidInput}, errs...)...)
	}
	return nil
}

func validateOptionalHTTPOrigin(field, value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" {
		return fmt.Errorf("%s must be a valid HTTP origin", field)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%s must use http or https", field)
	}
	if parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("%s must be an origin without path, query, fragment, or credentials", field)
	}
	return nil
}

func validateSMTP(settings SMTPSettings) error {
	var errs []error
	errs = append(errs, validateSMTPBasics(settings)...)
	errs = append(errs, validateSMTPConfiguredFields(settings)...)
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func validateSMTPBasics(settings SMTPSettings) []error {
	var errs []error
	if settings.Port < 1 || settings.Port > 65535 {
		errs = append(errs, errors.New("smtp.port must be between 1 and 65535"))
	}
	if settings.TimeoutSeconds <= 0 {
		errs = append(errs, errors.New("smtp.timeoutSeconds must be greater than 0"))
	}
	switch strings.TrimSpace(settings.TLSMode) {
	case "starttls", "implicit", "none":
	default:
		errs = append(errs, errors.New("smtp.tlsMode must be one of starttls, implicit, or none"))
	}
	return errs
}

func validateSMTPConfiguredFields(settings SMTPSettings) []error {
	if !smtpPartiallyConfigured(settings) {
		return nil
	}

	var errs []error
	if strings.TrimSpace(settings.Host) == "" {
		errs = append(errs, errors.New("smtp.host must not be empty when SMTP is configured"))
	}
	if err := validateSMTPFrom(settings.From); err != nil {
		errs = append(errs, err)
	}
	if (strings.TrimSpace(settings.Username) == "") != (settings.Password == "") {
		errs = append(errs, errors.New("smtp.username and smtp.password must be set together"))
	}
	return errs
}

func smtpPartiallyConfigured(settings SMTPSettings) bool {
	return strings.TrimSpace(settings.Host) != "" ||
		strings.TrimSpace(settings.Username) != "" ||
		settings.Password != "" ||
		strings.TrimSpace(settings.From) != ""
}

func smtpDeliveryConfigured(settings SMTPSettings) bool {
	return strings.TrimSpace(settings.Host) != "" && strings.TrimSpace(settings.From) != ""
}

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

func normalizeExportDataInput(input ExportDataInput) (ExportDataInput, error) {
	currentUserID, currentUserErr := identity.VNUserID(input.CurrentUserID)
	if currentUserErr == nil {
		return ExportDataInput{CurrentUserID: currentUserID}, nil
	}

	collector := appvalidation.Collector{}
	collector.AddError("currentUserId", currentUserErr, input.CurrentUserID)
	return ExportDataInput{}, collector.Err(ErrInvalidInput)
}

func normalizeImportDataInput(input ImportDataInput) (ImportDataInput, error) {
	currentUserID, currentUserErr := identity.VNUserID(input.CurrentUserID)
	if currentUserErr != nil {
		collector := appvalidation.Collector{}
		collector.AddError("currentUserId", currentUserErr, input.CurrentUserID)
		return ImportDataInput{}, collector.Err(ErrInvalidInput)
	}
	if strings.TrimSpace(input.Export.Format) == "" || input.Export.Tables == nil {
		return ImportDataInput{}, ErrDataImportInvalid
	}
	input.CurrentUserID = currentUserID
	return input, nil
}

func systemAdminAuditKey(admin domainsystem.AdminUser) string {
	return "system_admin:" + admin.ID
}

func managedUserAuditKey(user ManagedUser) string {
	return "user:" + user.ID
}

func validateSMTPFrom(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return errors.New("smtp.from must not be empty when SMTP is configured")
	}
	address, err := mail.ParseAddress(trimmed)
	if err != nil || address.Address != trimmed || address.Name != "" {
		return errors.New("smtp.from must be a valid email address")
	}
	return nil
}

func trimStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func jsonValue(value any) json.RawMessage {
	data, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return data
}
