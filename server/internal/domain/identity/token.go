package identity

type AccessTokenInput struct {
	Subject     string
	Email       string
	DisplayName *string
}

type IssuedToken struct {
	Value     string
	TokenType string
	ExpiresIn int
}

type AccessTokenClaims struct {
	Subject     string
	Email       string
	DisplayName *string
}
