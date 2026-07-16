package auth

import (
	"errors"
	"net/http"
	"time"

	appapitoken "github.com/yorukot/netstamp/internal/controller/application/apitoken"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

type createAPITokenBody struct {
	Name            string    `json:"name"`
	Scopes          []string  `json:"scopes"`
	ExpiresAt       time.Time `json:"expiresAt"`
	CurrentPassword string    `json:"currentPassword"`
}

type apiTokenBody struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	TokenHint  string     `json:"tokenHint"`
	Scopes     []string   `json:"scopes"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	ExpiresAt  time.Time  `json:"expiresAt"`
}

type createAPITokenResponse struct {
	Token apiTokenBody `json:"token"`
	Value string       `json:"value"`
}

type listAPITokensResponse struct {
	Tokens []apiTokenBody `json:"tokens"`
}

func (h *Handler) handleCreateAPIToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpmiddleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		httpx.WriteProblem(w, r, httpx.Unauthorized("missing authenticated user"))
		return
	}
	var body createAPITokenBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.apiTokens.Create(r.Context(), appapitoken.CreateInput{CurrentUserID: userID, Name: body.Name, Scopes: body.Scopes, ExpiresAt: body.ExpiresAt, CurrentPassword: body.CurrentPassword})
	if err != nil {
		httpx.WriteProblem(w, r, mapAPITokenError(err))
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httpx.WriteJSON(w, http.StatusCreated, createAPITokenResponse{Token: newAPITokenBody(output.Token), Value: output.RawToken})
}

func (h *Handler) handleListAPITokens(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpmiddleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		httpx.WriteProblem(w, r, httpx.Unauthorized("missing authenticated user"))
		return
	}
	tokens, err := h.apiTokens.List(r.Context(), appapitoken.ListInput{CurrentUserID: userID})
	if err != nil {
		httpx.WriteProblem(w, r, mapAPITokenError(err))
		return
	}
	items := make([]apiTokenBody, 0, len(tokens))
	for _, token := range tokens {
		items = append(items, newAPITokenBody(token))
	}
	w.Header().Set("Cache-Control", "no-store")
	httpx.WriteJSON(w, http.StatusOK, listAPITokensResponse{Tokens: items})
}

func (h *Handler) handleRevokeAPIToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpmiddleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		httpx.WriteProblem(w, r, httpx.Unauthorized("missing authenticated user"))
		return
	}
	err := h.apiTokens.Revoke(r.Context(), appapitoken.RevokeInput{CurrentUserID: userID, TokenID: httpx.Path(r, "token_id")})
	if err != nil {
		httpx.WriteProblem(w, r, mapAPITokenError(err))
		return
	}
	httpx.WriteNoContent(w)
}

func newAPITokenBody(token identity.APIToken) apiTokenBody {
	scopes := make([]string, 0, len(token.Scopes))
	for _, scope := range token.Scopes {
		scopes = append(scopes, string(scope))
	}
	return apiTokenBody{ID: token.ID, Name: token.Name, TokenHint: token.TokenHint, Scopes: scopes, CreatedAt: token.CreatedAt, LastUsedAt: token.LastUsedAt, ExpiresAt: token.ExpiresAt}
}

func mapAPITokenError(err error) error {
	switch {
	case errors.Is(err, appapitoken.ErrInvalidInput):
		return httpx.UnprocessableEntity("invalid api token input")
	case errors.Is(err, appapitoken.ErrCredentialsInvalid):
		return httpx.UnauthorizedCode(httpx.CodeAuthInvalidCredentials, "current password is invalid")
	case errors.Is(err, appapitoken.ErrTokenNotFound):
		return httpx.NotFoundCode(httpx.CodeAPITokenNotFound, "api token not found")
	case errors.Is(err, appapitoken.ErrTokenLimitReached):
		return httpx.ConflictCode(httpx.CodeAPITokenLimitReached, "active api token limit reached")
	default:
		return httpx.InternalServerError("api token operation failed")
	}
}
