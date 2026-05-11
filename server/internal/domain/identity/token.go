package identity

type IssuedToken struct {
	Value     string
	TokenType string
	ExpiresIn int
}

type AccessTokenClaims struct {
	Subject     string
	Email       string
}
