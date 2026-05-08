package auth

type RegisterInput struct {
	Email       string
	DisplayName string
	Password    string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthAccessResult struct {
	UserID      string
	Email       string
	DisplayName *string
	AccessToken string
	TokenType   string
	ExpiresIn   int
}
