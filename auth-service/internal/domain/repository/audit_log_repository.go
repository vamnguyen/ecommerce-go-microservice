package repository

import (
	"context"

	"auth-service/internal/domain/entity"

	"github.com/google/uuid"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.AuditLog, error)
	DeleteOlderThan(ctx context.Context, days int) error
}
