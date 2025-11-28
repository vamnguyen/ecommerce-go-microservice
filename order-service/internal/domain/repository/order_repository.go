package repository

import (
	"context"

	"order-service/internal/domain/entity"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Order, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Order, int64, error)
	Update(ctx context.Context, order *entity.Order) error
	List(ctx context.Context, limit, offset int) ([]*entity.Order, int64, error)
}

