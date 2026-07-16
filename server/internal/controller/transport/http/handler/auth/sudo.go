package auth

import (
	"errors"
	"net/http"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

func (h *Handler) handleSudoStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpmiddleware.SessionClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteProblem(w, r, httpx.UnauthorizedCode(httpx.CodeAuthMissingSession, "authentication required"))
		return
	}
	result, err := h.service.SudoStatus(r.Context(), claims.UserID, claims.SessionID)
	if err != nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("could not inspect recent authentication"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"active": result.Active, "expiresAt": result.ExpiresAt, "methods": result.Methods,
	})
}

func (h *Handler) handlePasswordSudo(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpmiddleware.SessionClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteProblem(w, r, httpx.UnauthorizedCode(httpx.CodeAuthMissingSession, "authentication required"))
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	if err := h.service.ReauthenticatePassword(r.Context(), claims.UserID, claims.SessionID, body.Password); err != nil {
		if errors.Is(err, appauth.ErrCredentialsInvalid) {
			httpx.WriteProblem(w, r, httpx.UnauthorizedCode(httpx.CodeAuthInvalidCredentials, "invalid password"))
			return
		}
		httpx.WriteProblem(w, r, httpx.InternalServerError("recent authentication failed"))
		return
	}
	httpx.WriteNoContent(w)
}
