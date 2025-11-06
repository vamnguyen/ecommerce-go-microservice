package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	domainErr "auth-service/internal/domain/errors"
	"auth-service/internal/domain/service"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTService(secret string, accessTokenTTL, refreshTokenTTL time.Duration) *JWTService {
	return &JWTService{
		secret:          []byte(secret),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (s *JWTService) GenerateAccessToken(claims service.TokenClaims) (string, error) {
	now := time.Now()
	jwtClaims := Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (s *JWTService) GenerateRefreshToken() (plainToken string, hashedToken string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}

	plainToken = base64.URLEncoding.EncodeToString(b)
	hashedToken = s.HashToken(plainToken)

	return plainToken, hashedToken, nil
}

func (s *JWTService) ValidateAccessToken(tokenString string) (*service.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, domainErr.ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, domainErr.ErrInvalidToken
	}

	var issuedAt int64
	if claims.IssuedAt != nil {
		issuedAt = claims.IssuedAt.Unix()
	}

	return &service.TokenClaims{
		UserID:   claims.UserID,
		Email:    claims.Email,
		Role:     claims.Role,
		IssuedAt: issuedAt,
	}, nil
}

func (s *JWTService) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}

func (s *JWTService) GetAccessTokenExpiry() time.Duration {
	return s.accessTokenTTL
}

func (s *JWTService) GetRefreshTokenExpiry() time.Duration {
	return s.refreshTokenTTL
}
