package admin

import (
	"net/http"
	"time"

	appadmin "github.com/yorukot/netstamp/internal/controller/application/admin"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

type grantSystemAdminBody struct {
	Email string `json:"email"`
}

type systemAdminsResponseBody struct {
	Admins []systemAdminResponseBody `json:"admins"`
}

type systemAdminResponseWrapper struct {
	Admin systemAdminResponseBody `json:"admin"`
}

type systemAdminResponseBody struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"displayName"`
	EmailVerified bool      `json:"emailVerified"`
	GrantedAt     time.Time `json:"grantedAt"`
}

func (h *Handler) handleListSystemAdmins(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("admin service is unavailable"))
		return
	}
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}

	admins, err := h.service.ListSystemAdmins(r.Context(), appadmin.ListSystemAdminsInput{CurrentUserID: userID})
	if err != nil {
		httpx.WriteProblem(w, r, mapAdminError(err, "list system administrators failed"))
		return
	}

	httpx.WriteJSON(w, http.StatusOK, systemAdminsResponse(admins))
}

func (h *Handler) handleGrantSystemAdmin(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("admin service is unavailable"))
		return
	}
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body grantSystemAdminBody
	if decodeErr := httpx.DecodeJSON(r, &body); decodeErr != nil {
		httpx.WriteProblem(w, r, decodeErr)
		return
	}

	admin, err := h.service.GrantSystemAdmin(r.Context(), appadmin.GrantSystemAdminInput{CurrentUserID: userID, Email: body.Email})
	if err != nil {
		httpx.WriteProblem(w, r, mapAdminError(err, "grant system administrator failed"))
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, systemAdminResponseWrapper{Admin: systemAdminResponse(admin)})
}

func (h *Handler) handleRevokeSystemAdmin(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("admin service is unavailable"))
		return
	}
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}

	err = h.service.RevokeSystemAdmin(r.Context(), appadmin.RevokeSystemAdminInput{CurrentUserID: userID, UserID: httpx.Path(r, "user_id")})
	if err != nil {
		httpx.WriteProblem(w, r, mapAdminError(err, "revoke system administrator failed"))
		return
	}

	httpx.WriteNoContent(w)
}

func systemAdminsResponse(admins []appadmin.SystemAdmin) systemAdminsResponseBody {
	response := systemAdminsResponseBody{Admins: make([]systemAdminResponseBody, 0, len(admins))}
	for _, admin := range admins {
		response.Admins = append(response.Admins, systemAdminResponse(admin))
	}
	return response
}

func systemAdminResponse(admin appadmin.SystemAdmin) systemAdminResponseBody {
	return systemAdminResponseBody{
		ID:            admin.ID,
		Email:         admin.Email,
		DisplayName:   admin.DisplayName,
		EmailVerified: admin.EmailVerifiedAt != nil,
		GrantedAt:     admin.GrantedAt,
	}
}
