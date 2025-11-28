package repository

import (
	"context"

	"user-service/internal/domain/entity"

	"github.com/google/uuid"
)

type UserProfileRepository interface {
	Create(ctx context.Context, profile *entity.UserProfile) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserProfile, error)
	Update(ctx context.Context, profile *entity.UserProfile) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.UserProfile, error)
	List(ctx context.Context, limit, offset int) ([]*entity.UserProfile, int64, error)
}

