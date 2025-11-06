package handler

import (
	"net/http"

	"auth-service/internal/application/dto"
	"auth-service/internal/application/usecase"
	domainErr "auth-service/internal/domain/errors"
	"auth-service/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
	logger      *logger.Logger
}

func NewAuthHandler(authUseCase *usecase.AuthUseCase, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	if err := h.authUseCase.Register(c.Request.Context(), req); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.MessageResponse{
		Message: "user registered successfully",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	result, err := h.authUseCase.Login(c.Request.Context(), req, ipAddress, userAgent)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	result, err := h.authUseCase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get("userID")
	userIDStr, _ := userID.(string)

	var req dto.RefreshTokenRequest
	_ = c.ShouldBindJSON(&req)

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.authUseCase.Logout(c.Request.Context(), userIDStr, req.RefreshToken, ipAddress, userAgent); err != nil {
		h.logger.Error("logout failed", zap.Error(err))
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID, _ := c.Get("userID")
	userIDStr, _ := userID.(string)

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.authUseCase.LogoutAll(c.Request.Context(), userIDStr, ipAddress, userAgent); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("userID")
	userIDStr, _ := userID.(string)

	user, err := h.authUseCase.GetMe(c.Request.Context(), userIDStr)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	userID, _ := c.Get("userID")
	userIDStr, _ := userID.(string)

	if err := h.authUseCase.ChangePassword(c.Request.Context(), userIDStr, req); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "password changed successfully",
	})
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	h.logger.Error("request failed", zap.Error(err))

	switch err {
	case domainErr.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	case domainErr.ErrUserAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
	case domainErr.ErrInvalidCredentials:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	case domainErr.ErrAccountLocked:
		c.JSON(http.StatusForbidden, gin.H{"error": "account is locked"})
	case domainErr.ErrAccountInactive:
		c.JSON(http.StatusForbidden, gin.H{"error": "account is inactive"})
	case domainErr.ErrInvalidToken, domainErr.ErrTokenExpired, domainErr.ErrTokenRevoked:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
	case domainErr.ErrWeakPassword:
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is too weak"})
	case domainErr.ErrInvalidPassword:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
