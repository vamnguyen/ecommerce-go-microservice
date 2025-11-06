package postgres

import (
	"context"
	"errors"
	"time"

	"auth-service/internal/domain/entity"
	domainErr "auth-service/internal/domain/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenModel struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash     string     `gorm:"uniqueIndex;not null"`
	TokenFamilyID *uuid.UUID `gorm:"type:uuid;index"`
	ExpiresAt     time.Time  `gorm:"not null;index"`
	IsRevoked     bool       `gorm:"not null;default:false"`
	CreatedAt     time.Time
	RevokedAt     *time.Time
}

func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	model := r.toModel(token)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	var model RefreshTokenModel
	if err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErr.ErrInvalidToken
		}
		return nil, domainErr.ErrDatabase
	}
	return r.toEntity(&model), nil
}

func (r *RefreshTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.RefreshToken, error) {
	var models []RefreshTokenModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_revoked = ? AND expires_at > ?", userID, false, time.Now()).
		Find(&models).Error; err != nil {
		return nil, domainErr.ErrDatabase
	}

	tokens := make([]*entity.RefreshToken, len(models))
	for i, model := range models {
		tokens[i] = r.toEntity(&model)
	}
	return tokens, nil
}

func (r *RefreshTokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&RefreshTokenModel{}).
		Where("token_hash = ?", tokenHash).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": now,
		}).Error; err != nil {
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *RefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&RefreshTokenModel{}).
		Where("user_id = ? AND is_revoked = ?", userID, false).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": now,
		}).Error; err != nil {
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *RefreshTokenRepository) RevokeByTokenFamilyID(ctx context.Context, familyID uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&RefreshTokenModel{}).
		Where("token_family_id = ? AND is_revoked = ?", familyID, false).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": now,
		}).Error; err != nil {
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	if err := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&RefreshTokenModel{}).Error; err != nil {
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *RefreshTokenRepository) toModel(token *entity.RefreshToken) *RefreshTokenModel {
	return &RefreshTokenModel{
		ID:            token.ID,
		UserID:        token.UserID,
		TokenHash:     token.TokenHash,
		TokenFamilyID: &token.TokenFamilyID,
		ExpiresAt:     token.ExpiresAt,
		IsRevoked:     token.IsRevoked,
		CreatedAt:     token.CreatedAt,
		RevokedAt:     token.RevokedAt,
	}
}

func (r *RefreshTokenRepository) toEntity(model *RefreshTokenModel) *entity.RefreshToken {
	var familyID uuid.UUID
	if model.TokenFamilyID != nil {
		familyID = *model.TokenFamilyID
	}
	return &entity.RefreshToken{
		ID:            model.ID,
		UserID:        model.UserID,
		TokenHash:     model.TokenHash,
		TokenFamilyID: familyID,
		ExpiresAt:     model.ExpiresAt,
		IsRevoked:     model.IsRevoked,
		CreatedAt:     model.CreatedAt,
		RevokedAt:     model.RevokedAt,
	}
}
