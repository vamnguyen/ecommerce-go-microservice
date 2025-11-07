package usecase

import (
	"context"
	"time"

	"auth-service/internal/application/dto"
	"auth-service/internal/domain/entity"
	domainErr "auth-service/internal/domain/errors"
	"auth-service/internal/domain/repository"
	"auth-service/internal/domain/service"

	"github.com/google/uuid"
)

type AuthUseCase struct {
	userRepo           repository.UserRepository
	refreshTokenRepo   repository.RefreshTokenRepository
	tokenBlacklistRepo repository.TokenBlacklistRepository
	auditLogRepo       repository.AuditLogRepository
	passwordService    service.PasswordService
	tokenService       service.TokenService
	config             AuthConfig
}

type AuthConfig struct {
	MaxLoginAttempts    int
	AccountLockDuration time.Duration
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	tokenBlacklistRepo repository.TokenBlacklistRepository,
	auditLogRepo repository.AuditLogRepository,
	passwordService service.PasswordService,
	tokenService service.TokenService,
	config AuthConfig,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:           userRepo,
		refreshTokenRepo:   refreshTokenRepo,
		tokenBlacklistRepo: tokenBlacklistRepo,
		auditLogRepo:       auditLogRepo,
		passwordService:    passwordService,
		tokenService:       tokenService,
		config:             config,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, req dto.RegisterRequest) error {
	exists, err := uc.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return domainErr.ErrDatabase
	}
	if exists {
		return domainErr.ErrUserAlreadyExists
	}

	if err := uc.passwordService.ValidatePasswordStrength(req.Password); err != nil {
		return err
	}

	passwordHash, err := uc.passwordService.HashPassword(req.Password)
	if err != nil {
		return domainErr.ErrInternalServer
	}

	user := entity.NewUser(req.Email, passwordHash)

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return domainErr.ErrDatabase
	}

	auditLog := entity.NewAuditLog(user.ID, entity.AuditActionRegister, "", "")
	_ = uc.auditLogRepo.Create(ctx, auditLog)

	return nil
}

func (uc *AuthUseCase) Login(ctx context.Context, req dto.LoginRequest, ipAddress, userAgent string) (*dto.AuthResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		auditLog := entity.NewAuditLog(uuid.Nil, entity.AuditActionLoginFailed, ipAddress, userAgent)
		auditLog.AddMetadata("email", req.Email)
		_ = uc.auditLogRepo.Create(ctx, auditLog)
		return nil, domainErr.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, domainErr.ErrAccountInactive
	}

	if user.IsAccountLocked() {
		auditLog := entity.NewAuditLog(user.ID, entity.AuditActionAccountLocked, ipAddress, userAgent)
		_ = uc.auditLogRepo.Create(ctx, auditLog)
		return nil, domainErr.ErrAccountLocked
	}

	if err := uc.passwordService.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		user.IncrementFailedLoginAttempts(uc.config.MaxLoginAttempts, uc.config.AccountLockDuration)
		_ = uc.userRepo.Update(ctx, user)

		auditLog := entity.NewAuditLog(user.ID, entity.AuditActionLoginFailed, ipAddress, userAgent)
		_ = uc.auditLogRepo.Create(ctx, auditLog)

		return nil, domainErr.ErrInvalidCredentials
	}

	user.ResetFailedLoginAttempts()
	user.UpdateLastLogin(ipAddress)
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, domainErr.ErrDatabase
	}

	claims := service.TokenClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   string(user.Role),
	}

	accessToken, err := uc.tokenService.GenerateAccessToken(claims)
	if err != nil {
		return nil, domainErr.ErrInternalServer
	}

	refreshPlain, refreshHash, err := uc.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, domainErr.ErrInternalServer
	}

	expiresAt := time.Now().Add(uc.tokenService.GetRefreshTokenExpiry())
	refreshToken := entity.NewRefreshToken(user.ID, refreshHash, expiresAt)
	if err := uc.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, domainErr.ErrDatabase
	}

	auditLog := entity.NewAuditLog(user.ID, entity.AuditActionLogin, ipAddress, userAgent)
	_ = uc.auditLogRepo.Create(ctx, auditLog)

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshPlain,
	}, nil
}

func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshPlain, accessToken, ipAddress, userAgent string) (*dto.RefreshTokenResponse, error) {
	if refreshPlain == "" {
		return nil, domainErr.ErrMissingToken
	}

	refreshHash := uc.tokenService.HashToken(refreshPlain)

	token, err := uc.refreshTokenRepo.FindByTokenHash(ctx, refreshHash)
	if err != nil {
		return nil, domainErr.ErrInvalidToken
	}

	if !token.IsValid() {
		return nil, domainErr.ErrInvalidToken
	}

	user, err := uc.userRepo.FindByID(ctx, token.UserID)
	if err != nil {
		return nil, domainErr.ErrUserNotFound
	}

	// Revoke the used refresh token
	if err := uc.refreshTokenRepo.RevokeByTokenHash(ctx, refreshHash); err != nil {
		return nil, domainErr.ErrDatabase
	}

	// Blacklist the access token
	if accessToken != "" {
		accessHash := uc.tokenService.HashToken(accessToken)
		expiresAt := time.Now().Add(uc.tokenService.GetAccessTokenExpiry())
		blacklist := entity.NewTokenBlacklist(accessHash, expiresAt)
		_ = uc.tokenBlacklistRepo.Add(ctx, blacklist)
	}

	claims := service.TokenClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   string(user.Role),
	}

	newAccessToken, err := uc.tokenService.GenerateAccessToken(claims)
	if err != nil {
		return nil, domainErr.ErrInternalServer
	}

	newRefreshPlain, newRefreshHash, err := uc.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, domainErr.ErrInternalServer
	}

	expiresAt := time.Now().Add(uc.tokenService.GetRefreshTokenExpiry())
	newRefreshToken := entity.NewRefreshTokenWithFamily(user.ID, newRefreshHash, expiresAt, token.TokenFamilyID)
	if err := uc.refreshTokenRepo.Create(ctx, newRefreshToken); err != nil {
		return nil, domainErr.ErrDatabase
	}

	auditLog := entity.NewAuditLog(user.ID, entity.AuditActionTokenRefresh, "", "")
	_ = uc.auditLogRepo.Create(ctx, auditLog)

	return &dto.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshPlain,
	}, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, userID string, refreshPlain, accessToken, ipAddress, userAgent string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return domainErr.ErrInvalidInput
	}

	// Revoke the provided refresh token
	if refreshPlain != "" {
		refreshHash := uc.tokenService.HashToken(refreshPlain)
		_ = uc.refreshTokenRepo.RevokeByTokenHash(ctx, refreshHash)
	}

	// Blacklist the access token
	if accessToken != "" {
		accessHash := uc.tokenService.HashToken(accessToken)
		expiresAt := time.Now().Add(uc.tokenService.GetAccessTokenExpiry())
		blacklist := entity.NewTokenBlacklist(accessHash, expiresAt)
		_ = uc.tokenBlacklistRepo.Add(ctx, blacklist)
	}

	auditLog := entity.NewAuditLog(userUUID, entity.AuditActionLogout, ipAddress, userAgent)
	_ = uc.auditLogRepo.Create(ctx, auditLog)

	return nil
}

func (uc *AuthUseCase) LogoutAll(ctx context.Context, userID, accessToken, ipAddress, userAgent string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return domainErr.ErrInvalidInput
	}

	// Revoke all refresh tokens of the user
	if err := uc.refreshTokenRepo.RevokeAllByUserID(ctx, userUUID); err != nil {
		return domainErr.ErrDatabase
	}

	// Blacklist the access token
	if accessToken != "" {
		accessHash := uc.tokenService.HashToken(accessToken)
		expiresAt := time.Now().Add(uc.tokenService.GetAccessTokenExpiry())
		blacklist := entity.NewTokenBlacklist(accessHash, expiresAt)
		_ = uc.tokenBlacklistRepo.Add(ctx, blacklist)
	}

	auditLog := entity.NewAuditLog(userUUID, entity.AuditActionLogout, ipAddress, userAgent)
	auditLog.AddMetadata("logout_all", true)
	_ = uc.auditLogRepo.Create(ctx, auditLog)

	return nil
}

func (uc *AuthUseCase) GetMe(ctx context.Context, userID string) (*dto.UserDTO, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, domainErr.ErrInvalidInput
	}

	user, err := uc.userRepo.FindByID(ctx, userUUID)
	if err != nil {
		return nil, domainErr.ErrUserNotFound
	}

	return &dto.UserDTO{
		ID:         user.ID.String(),
		Email:      user.Email,
		Role:       string(user.Role),
		IsVerified: user.IsVerified,
		IsActive:   user.IsActive,
		CreatedAt:  user.CreatedAt,
	}, nil
}

func (uc *AuthUseCase) ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return domainErr.ErrInvalidInput
	}

	user, err := uc.userRepo.FindByID(ctx, userUUID)
	if err != nil {
		return domainErr.ErrUserNotFound
	}

	if err := uc.passwordService.VerifyPassword(user.PasswordHash, req.OldPassword); err != nil {
		return domainErr.ErrInvalidPassword
	}

	if err := uc.passwordService.ValidatePasswordStrength(req.NewPassword); err != nil {
		return err
	}

	newPasswordHash, err := uc.passwordService.HashPassword(req.NewPassword)
	if err != nil {
		return domainErr.ErrInternalServer
	}

	user.UpdatePassword(newPasswordHash)

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return domainErr.ErrDatabase
	}

	_ = uc.refreshTokenRepo.RevokeAllByUserID(ctx, userUUID)

	auditLog := entity.NewAuditLog(user.ID, entity.AuditActionPasswordChange, "", "")
	_ = uc.auditLogRepo.Create(ctx, auditLog)

	return nil
}
