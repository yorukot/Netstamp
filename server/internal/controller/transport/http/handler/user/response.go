package userhttp

import "github.com/yorukot/netstamp/internal/domain/identity"

type userOutput struct {
	Body userOutputBody
}

type userOutputBody struct {
	User userResponse `json:"user"`
}

type userResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"displayName"`
	HasPassword   bool   `json:"hasPassword"`
	EmailVerified bool   `json:"emailVerified"`
	IsSystemAdmin bool   `json:"isSystemAdmin"`
}

func newUserOutput(user identity.User) *userOutput {
	return &userOutput{
		Body: userOutputBody{
			User: userResponse{
				ID:            user.ID,
				Email:         user.Email,
				DisplayName:   user.DisplayName,
				HasPassword:   user.HasPassword,
				EmailVerified: user.EmailVerifiedAt != nil,
				IsSystemAdmin: user.IsSystemAdmin,
			},
		},
	}
}
