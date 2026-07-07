package auth

type userResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"displayName"`
	IsSystemAdmin bool   `json:"isSystemAdmin"`
}
