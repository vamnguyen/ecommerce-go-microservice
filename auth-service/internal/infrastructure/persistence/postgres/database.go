package postgres

import (
	"fmt"
	"time"

	"auth-service/internal/infrastructure/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&UserModel{},
		&RefreshTokenModel{},
		&TokenBlacklistModel{},
		&AuditLogModel{},
	)
}

type BaseModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}
