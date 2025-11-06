package repository

import (
	"context"

	"auth-service/internal/domain/entity"
)

type TokenBlacklistRepository interface {
	Add(ctx context.Context, blacklist *entity.TokenBlacklist) error
	IsBlacklisted(ctx context.Context, tokenHash string) (bool, error)
	DeleteExpired(ctx context.Context) error
}
