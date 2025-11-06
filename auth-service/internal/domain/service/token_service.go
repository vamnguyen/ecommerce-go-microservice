package service

import "time"

type TokenClaims struct {
	UserID string
	Email  string
	Role   string
}

type TokenService interface {
	GenerateAccessToken(claims TokenClaims) (string, error)
	GenerateRefreshToken() (plainToken string, hashedToken string, err error)
	ValidateAccessToken(token string) (*TokenClaims, error)
	HashToken(token string) string
	GetAccessTokenExpiry() time.Duration
	GetRefreshTokenExpiry() time.Duration
}
