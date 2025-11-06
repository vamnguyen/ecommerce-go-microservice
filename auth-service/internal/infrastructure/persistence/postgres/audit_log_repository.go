package postgres

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"

	"auth-service/internal/domain/entity"
	domainErr "auth-service/internal/domain/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

type AuditLogModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Action    string    `gorm:"not null;index"`
	IPAddress string
	UserAgent string
	Metadata  JSONB     `gorm:"type:jsonb"`
	CreatedAt time.Time `gorm:"index"`
}

func (AuditLogModel) TableName() string {
	return "audit_logs"
}

type AuditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	model := r.toModel(log)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *AuditLogRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.AuditLog, error) {
	var models []AuditLogModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error; err != nil {
		return nil, domainErr.ErrDatabase
	}

	logs := make([]*entity.AuditLog, len(models))
	for i, model := range models {
		logs[i] = r.toEntity(&model)
	}
	return logs, nil
}

func (r *AuditLogRepository) DeleteOlderThan(ctx context.Context, days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	if err := r.db.WithContext(ctx).
		Where("created_at < ?", cutoffDate).
		Delete(&AuditLogModel{}).Error; err != nil {
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *AuditLogRepository) toModel(log *entity.AuditLog) *AuditLogModel {
	return &AuditLogModel{
		ID:        log.ID,
		UserID:    log.UserID,
		Action:    string(log.Action),
		IPAddress: log.IPAddress,
		UserAgent: log.UserAgent,
		Metadata:  JSONB(log.Metadata),
		CreatedAt: log.CreatedAt,
	}
}

func (r *AuditLogRepository) toEntity(model *AuditLogModel) *entity.AuditLog {
	return &entity.AuditLog{
		ID:        model.ID,
		UserID:    model.UserID,
		Action:    entity.AuditAction(model.Action),
		IPAddress: model.IPAddress,
		UserAgent: model.UserAgent,
		Metadata:  map[string]interface{}(model.Metadata),
		CreatedAt: model.CreatedAt,
	}
}
