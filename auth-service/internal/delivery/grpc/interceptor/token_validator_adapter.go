package interceptor

import "auth-service/internal/domain/service"

type TokenServiceAdapter struct {
	tokenService service.TokenService
}

func NewTokenServiceAdapter(tokenService service.TokenService) *TokenServiceAdapter {
	return &TokenServiceAdapter{
		tokenService: tokenService,
	}
}

func (a *TokenServiceAdapter) ValidateAccessToken(token string) (*TokenClaims, error) {
	claims, err := a.tokenService.ValidateAccessToken(token)
	if err != nil {
		return nil, err
	}

	return &TokenClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}
