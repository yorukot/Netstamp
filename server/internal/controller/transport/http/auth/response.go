package auth

type userResponse struct {
	ID          string `json:"id" format:"uuid" doc:"User UUID." example:"11111111-1111-1111-1111-111111111111"`
	Email       string `json:"email" format:"email" doc:"Normalized email address used to sign in." example:"user@example.com"`
	DisplayName string `json:"displayName" minLength:"1" maxLength:"64" doc:"Name shown in the app." example:"Jane Doe"`
}
