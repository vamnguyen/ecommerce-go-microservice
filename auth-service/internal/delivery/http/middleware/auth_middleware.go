package middleware

import (
	"net/http"
	"strings"

	"auth-service/internal/domain/repository"
	"auth-service/internal/domain/service"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	tokenService       service.TokenService
	tokenBlacklistRepo repository.TokenBlacklistRepository
	userRepo           repository.UserRepository
}

func NewAuthMiddleware(tokenService service.TokenService, tokenBlacklistRepo repository.TokenBlacklistRepository, userRepo repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService:       tokenService,
		tokenBlacklistRepo: tokenBlacklistRepo,
		userRepo:           userRepo,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := m.tokenService.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		tokenHash := m.tokenService.HashToken(token)
		isBlacklisted, err := m.tokenBlacklistRepo.IsBlacklisted(c.Request.Context(), tokenHash)
		if err == nil && isBlacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("accessToken", token)

		c.Next()
	}
}
