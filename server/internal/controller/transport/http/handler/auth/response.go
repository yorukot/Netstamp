package auth

type userResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"displayName"`
	EmailVerified bool   `json:"emailVerified"`
	IsSystemAdmin bool   `json:"isSystemAdmin"`
}
