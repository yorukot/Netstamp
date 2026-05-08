package auth

type RegisterInput struct {
	Email       string
	DisplayName string
	Password    string //nolint:gosec // This DTO intentionally carries a plaintext password before hashing.
}

type LoginInput struct {
	Email    string
	Password string //nolint:gosec // This DTO intentionally carries a plaintext password for verification.
}

type AuthAccessResult struct {
	UserID      string
	Email       string
	DisplayName *string
	AccessToken string //nolint:gosec // This DTO intentionally carries the issued access token to the caller.
	TokenType   string
	ExpiresIn   int
}
