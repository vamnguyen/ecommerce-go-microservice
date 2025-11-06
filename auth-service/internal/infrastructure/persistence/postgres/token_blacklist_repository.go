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

type TokenBlacklistModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	TokenHash string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
}

func (TokenBlacklistModel) TableName() string {
	return "token_blacklist"
}

type TokenBlacklistRepository struct {
	db *gorm.DB
}

func NewTokenBlacklistRepository(db *gorm.DB) *TokenBlacklistRepository {
	return &TokenBlacklistRepository{db: db}
}

func (r *TokenBlacklistRepository) Add(ctx context.Context, blacklist *entity.TokenBlacklist) error {
	model := r.toModel(blacklist)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *TokenBlacklistRepository) IsBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&TokenBlacklistModel{}).
		Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).
		Count(&count).Error; err != nil {
		return false, domainErr.ErrDatabase
	}
	return count > 0, nil
}

func (r *TokenBlacklistRepository) DeleteExpired(ctx context.Context) error {
	if err := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&TokenBlacklistModel{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *TokenBlacklistRepository) toModel(blacklist *entity.TokenBlacklist) *TokenBlacklistModel {
	return &TokenBlacklistModel{
		ID:        blacklist.ID,
		TokenHash: blacklist.TokenHash,
		ExpiresAt: blacklist.ExpiresAt,
		CreatedAt: blacklist.CreatedAt,
	}
}
