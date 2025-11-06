package repository

import (
	"context"

	"auth-service/internal/domain/entity"

	"github.com/google/uuid"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.RefreshToken, error)
	RevokeByTokenHash(ctx context.Context, tokenHash string) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
	RevokeByTokenFamilyID(ctx context.Context, familyID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}
