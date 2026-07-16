package admin

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	appadmin "github.com/yorukot/netstamp/internal/controller/application/admin"
	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

type Handler struct {
	service    *appadmin.Service
	verifier   appauth.SessionManager
	cookieName string
	sudo       *appauth.Service
}

func NewHandler(service *appadmin.Service, verifier appauth.SessionManager, cookieName string, sudoServices ...*appauth.Service) *Handler {
	handler := &Handler{service: service, verifier: verifier, cookieName: cookieName}
	if len(sudoServices) > 0 {
		handler.sudo = sudoServices[0]
	}
	return handler
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireAuth(h.verifier, h.cookieName))

		r.Get("/admin/system-admins", h.handleListSystemAdmins)
		r.Get("/admin/users", h.handleListManagedUsers)
		r.Get("/admin/settings", h.handleGetSettings)
		registerSensitive := func(sensitive chi.Router) {
			sensitive.Post("/admin/system-admins", h.handleGrantSystemAdmin)
			sensitive.Delete("/admin/system-admins/{user_id}", h.handleRevokeSystemAdmin)
			sensitive.Patch("/admin/users/{user_id}", h.handleUpdateManagedUser)
			sensitive.Post("/admin/users/{user_id}/password", h.handleSetManagedUserPassword)
			sensitive.Delete("/admin/users/{user_id}/password", h.handleClearManagedUserPassword)
			sensitive.Get("/admin/data-export", h.handleExportData)
			sensitive.Post("/admin/data-import", h.handleImportData)
			sensitive.Patch("/admin/settings", h.handleUpdateSettings)
		}
		if h.sudo != nil {
			r.Group(func(sensitive chi.Router) {
				sensitive.Use(httpmiddleware.RequireSudo(h.sudo))
				registerSensitive(sensitive)
			})
		} else {
			registerSensitive(r)
		}
	})
}

func (h *Handler) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("admin settings service is unavailable"))
		return
	}
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	settings, err := h.service.GetSettings(r.Context(), appadmin.GetSettingsInput{CurrentUserID: userID})
	if err != nil {
		httpx.WriteProblem(w, r, mapAdminError(err, "get admin settings failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"settings": settingsResponse(settings)})
}

func (h *Handler) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("admin settings service is unavailable"))
		return
	}
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body settingsBody
	if decodeErr := httpx.DecodeJSON(r, &body); decodeErr != nil {
		httpx.WriteProblem(w, r, decodeErr)
		return
	}
	settings, err := h.service.UpdateSettings(r.Context(), body.updateInput(userID))
	if err != nil {
		httpx.WriteProblem(w, r, mapAdminError(err, "update admin settings failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"settings": settingsResponse(settings)})
}

func currentUserID(r *http.Request) (string, error) {
	userID, ok := httpmiddleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		return "", httpx.UnauthorizedCode(httpx.CodeAuthMissingSession, "missing auth session")
	}
	return userID, nil
}

func mapAdminError(err error, fallback string) error {
	switch {
	case errors.Is(err, appadmin.ErrForbidden):
		return httpx.ForbiddenCode(httpx.CodeSystemAdminRequired, "system administrator access is required")
	case errors.Is(err, identity.ErrUserNotFound), errors.Is(err, appadmin.ErrSystemAdminNotFound):
		return httpx.NotFoundCode(httpx.CodeUserNotFound, "user not found")
	case errors.Is(err, appadmin.ErrLastSystemAdmin):
		return httpx.ConflictCode(httpx.CodeLastSystemAdmin, "system must keep at least one administrator")
	case errors.Is(err, identity.ErrLastAuthenticationMethod):
		return httpx.ConflictCode(httpx.CodeAuthLastCredential, "account must keep at least one authentication method")
	case errors.Is(err, appadmin.ErrSelfSystemAdminRemoval), errors.Is(err, appadmin.ErrSelfAccountDisable):
		return httpx.ConflictCode(httpx.CodeSelfSystemAdminAction, "system administrator cannot remove or disable self")
	case errors.Is(err, appadmin.ErrDataImportInvalid):
		return httpx.UnprocessableEntityCode(httpx.CodeInvalidAdminDataImport, "invalid admin data import")
	case errors.Is(err, appadmin.ErrInvalidInput):
		return invalidAdminInputError(err)
	default:
		return httpx.InternalServerError(fallback)
	}
}

func invalidAdminInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return httpx.UnprocessableEntity("invalid admin input")
	}

	details := make([]httpx.ErrorDetail, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, httpx.ErrorDetail{
			Code:     fieldErr.Code,
			Message:  fieldErr.Message,
			Location: adminErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return httpx.UnprocessableEntity("invalid admin input", details...)
}

func adminErrorLocation(field string) string {
	switch field {
	case "userId":
		return "path.user_id"
	case "email":
		return "body.email"
	case "password":
		return "body.password"
	case "currentUserId":
		return "session.user.id"
	default:
		return "body." + field
	}
}
