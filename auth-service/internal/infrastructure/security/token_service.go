package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"auth-service/internal/domain/service"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	privateKey      *rsa.PrivateKey
	publicKey       *rsa.PublicKey
	publicKeyPEM    string
	algorithm       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTService(algorithm, privateKeyPath, publicKeyPath string, accessTokenTTL, refreshTokenTTL time.Duration) (*TokenService, error) {
	// Load private key
	privateKeyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key PEM")
	}

	var privateKey *rsa.PrivateKey
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		privateKey = key
	} else if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		privateKey = key.(*rsa.PrivateKey)
	} else {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Load public key
	publicKeyData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	block, _ = pem.Decode(publicKeyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode public key PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return &TokenService{
		privateKey:      privateKey,
		publicKey:       publicKey,
		publicKeyPEM:    string(publicKeyData),
		algorithm:       algorithm,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}, nil
}

func (s *TokenService) GenerateAccessToken(claims service.TokenClaims) (string, error) {
	now := time.Now()
	jwtClaims := Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "auth-service",
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims)
	signedToken, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (s *TokenService) GenerateRefreshToken() (plainToken string, hashedToken string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}

	plainToken = base64.URLEncoding.EncodeToString(b)
	hashedToken = s.HashToken(plainToken)

	return plainToken, hashedToken, nil
}

// func (s *TokenService) ExtractClaimsWithoutValidation(tokenString string) (*service.TokenClaims, error) {
// 	// Parse without validation
// 	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
// 	if err != nil {
// 		return nil, domainErr.ErrInvalidToken
// 	}

// 	claims, ok := token.Claims.(*Claims)
// 	if !ok || !token.Valid {
// 		return nil, domainErr.ErrInvalidToken
// 	}

// 	var issuedAt int64
// 	if claims.IssuedAt != nil {
// 		issuedAt = claims.IssuedAt.Unix()
// 	}

// 	return &service.TokenClaims{
// 		UserID:   claims.UserID,
// 		Email:    claims.Email,
// 		Role:     claims.Role,
// 		IssuedAt: issuedAt,
// 	}, nil
// }

func (s *TokenService) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}

func (s *TokenService) GetAccessTokenExpiry() time.Duration {
	return s.accessTokenTTL
}

func (s *TokenService) GetRefreshTokenExpiry() time.Duration {
	return s.refreshTokenTTL
}

func (s *TokenService) GetPublicKey() string {
	return s.publicKeyPEM
}

func (s *TokenService) GetAlgorithm() string {
	return s.algorithm
}

// ExtractClaimsWithoutValidation extracts claims without signature validation
// Used when JWT is already validated by Kong Gateway
func (s *TokenService) ExtractClaimsWithoutValidation(tokenString string) (*service.TokenClaims, error) {
	// Parse token without verification
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	// Return claims even if expired (Kong already validated)
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
