package handler

import (
	"net/http"

	"auth-service/internal/application/dto"
	"auth-service/internal/application/usecase"
	domainErr "auth-service/internal/domain/errors"
	"auth-service/internal/infrastructure/config"
	"auth-service/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
	logger      *logger.Logger
	config      *config.Config
}

func NewAuthHandler(authUseCase *usecase.AuthUseCase, logger *logger.Logger, config *config.Config) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		logger:      logger,
		config:      config,
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

	h.setRefreshTokenCookie(c, result.RefreshToken)

	response := dto.AuthResponse{
		AccessToken: result.AccessToken,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie(h.config.Cookie.RefreshTokenName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	result, err := h.authUseCase.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		h.clearRefreshTokenCookie(c)
		h.handleError(c, err)
		return
	}

	h.setRefreshTokenCookie(c, result.RefreshToken)

	response := dto.RefreshTokenResponse{
		AccessToken: result.AccessToken,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get("userID")
	userIDStr, _ := userID.(string)

	accessToken, _ := c.Get("accessToken")
	accessTokenStr, _ := accessToken.(string)

	refreshToken, _ := c.Cookie(h.config.Cookie.RefreshTokenName)

	if err := h.authUseCase.Logout(c.Request.Context(), userIDStr, refreshToken, accessTokenStr, c.ClientIP(), c.GetHeader("User-Agent")); err != nil {
		h.logger.Error("logout failed", zap.Error(err))
	}

	h.clearRefreshTokenCookie(c)

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID, _ := c.Get("userID")
	userIDStr, _ := userID.(string)

	accessToken, _ := c.Get("accessToken")
	accessTokenStr, _ := accessToken.(string)

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.authUseCase.LogoutAll(c.Request.Context(), userIDStr, accessTokenStr, ipAddress, userAgent); err != nil {
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

// Helpers

func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, refreshToken string) {
	c.SetCookie(
		h.config.Cookie.RefreshTokenName,
		refreshToken,
		int(h.config.JWT.RefreshTokenTTL.Seconds()),
		"/",
		h.config.Cookie.Domain,
		h.config.Cookie.Secure,
		true,
	)
}

func (h *AuthHandler) clearRefreshTokenCookie(c *gin.Context) {
	c.SetCookie(
		h.config.Cookie.RefreshTokenName,
		"",
		-1,
		"/",
		h.config.Cookie.Domain,
		h.config.Cookie.Secure,
		true,
	)
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
