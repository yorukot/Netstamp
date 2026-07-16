package admin

import (
	"net/http"
	"time"

	appadmin "github.com/yorukot/netstamp/internal/controller/application/admin"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

type managedUsersResponseBody struct {
	Users []managedUserResponseBody `json:"users"`
}

type managedUserResponseWrapper struct {
	User managedUserResponseBody `json:"user"`
}

type managedUserResponseBody struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	DisplayName   string     `json:"displayName"`
	EmailVerified bool       `json:"emailVerified"`
	DisabledAt    *time.Time `json:"disabledAt,omitempty"`
	IsSystemAdmin bool       `json:"isSystemAdmin"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	GrantedAt     *time.Time `json:"grantedAt,omitempty"`
	HasPassword   bool       `json:"hasPassword"`
}

type updateManagedUserBody struct {
	Disabled    *bool `json:"disabled,omitempty"`
	SystemAdmin *bool `json:"systemAdmin,omitempty"`
}

type setManagedUserPasswordBody struct {
	Password string `json:"password"`
}

func (h *Handler) handleListManagedUsers(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("admin service is unavailable"))
		return
	}
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}

	users, err := h.service.ListManagedUsers(r.Context(), appadmin.ListManagedUsersInput{CurrentUserID: userID})
	if err != nil {
		httpx.WriteProblem(w, r, mapAdminError(err, "list managed users failed"))
		return
	}

	httpx.WriteJSON(w, http.StatusOK, managedUsersResponse(users))
}

func (h *Handler) handleUpdateManagedUser(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("admin service is unavailable"))
		return
	}
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body updateManagedUserBody
	if decodeErr := httpx.DecodeJSON(r, &body); decodeErr != nil {
		httpx.WriteProblem(w, r, decodeErr)
		return
	}

	user, err := h.service.UpdateManagedUser(r.Context(), appadmin.UpdateManagedUserInput{
		CurrentUserID: userID,
		UserID:        httpx.Path(r, "user_id"),
		Disabled:      body.Disabled,
		SystemAdmin:   body.SystemAdmin,
	})
	if err != nil {
		httpx.WriteProblem(w, r, mapAdminError(err, "update managed user failed"))
		return
	}

	httpx.WriteJSON(w, http.StatusOK, managedUserResponseWrapper{User: managedUserResponse(user)})
}

func (h *Handler) handleSetManagedUserPassword(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("admin service is unavailable"))
		return
	}
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body setManagedUserPasswordBody
	if decodeErr := httpx.DecodeJSON(r, &body); decodeErr != nil {
		httpx.WriteProblem(w, r, decodeErr)
		return
	}

	user, err := h.service.SetManagedUserPassword(r.Context(), appadmin.SetManagedUserPasswordInput{
		CurrentUserID: userID,
		UserID:        httpx.Path(r, "user_id"),
		Password:      body.Password,
	})
	if err != nil {
		httpx.WriteProblem(w, r, mapAdminError(err, "set managed user password failed"))
		return
	}

	httpx.WriteJSON(w, http.StatusOK, managedUserResponseWrapper{User: managedUserResponse(user)})
}

func (h *Handler) handleClearManagedUserPassword(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("admin service is unavailable"))
		return
	}
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	user, err := h.service.ClearManagedUserPassword(r.Context(), appadmin.ClearManagedUserPasswordInput{CurrentUserID: userID, UserID: httpx.Path(r, "user_id")})
	if err != nil {
		httpx.WriteProblem(w, r, mapAdminError(err, "clear managed user password failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, managedUserResponseWrapper{User: managedUserResponse(user)})
}

func managedUsersResponse(users []appadmin.ManagedUser) managedUsersResponseBody {
	response := managedUsersResponseBody{Users: make([]managedUserResponseBody, 0, len(users))}
	for _, user := range users {
		response.Users = append(response.Users, managedUserResponse(user))
	}
	return response
}

func managedUserResponse(user appadmin.ManagedUser) managedUserResponseBody {
	return managedUserResponseBody{
		ID:            user.ID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		EmailVerified: user.EmailVerifiedAt != nil,
		DisabledAt:    user.DisabledAt,
		IsSystemAdmin: user.IsSystemAdmin,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		GrantedAt:     user.GrantedAt,
		HasPassword:   user.HasPassword,
	}
}
