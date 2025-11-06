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

type UserModel struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email               string    `gorm:"uniqueIndex;not null"`
	PasswordHash        string    `gorm:"not null"`
	Role                string    `gorm:"not null;default:'user'"`
	IsVerified          bool      `gorm:"not null;default:false"`
	IsActive            bool      `gorm:"not null;default:true"`
	FailedLoginAttempts int       `gorm:"not null;default:0"`
	LockedUntil         *time.Time
	LastLoginAt         *time.Time
	LastLoginIP         string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (UserModel) TableName() string {
	return "users"
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	model := r.toModel(user)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErr.ErrUserNotFound
		}
		return nil, domainErr.ErrDatabase
	}
	return r.toEntity(&model), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErr.ErrUserNotFound
		}
		return nil, domainErr.ErrDatabase
	}
	return r.toEntity(&model), nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	model := r.toModel(user)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&UserModel{}, "id = ?", id).Error; err != nil {
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, domainErr.ErrDatabase
	}
	return count > 0, nil
}

func (r *UserRepository) toModel(user *entity.User) *UserModel {
	return &UserModel{
		ID:                  user.ID,
		Email:               user.Email,
		PasswordHash:        user.PasswordHash,
		Role:                string(user.Role),
		IsVerified:          user.IsVerified,
		IsActive:            user.IsActive,
		FailedLoginAttempts: user.FailedLoginAttempts,
		LockedUntil:         user.LockedUntil,
		LastLoginAt:         user.LastLoginAt,
		LastLoginIP:         user.LastLoginIP,
		CreatedAt:           user.CreatedAt,
		UpdatedAt:           user.UpdatedAt,
	}
}

func (r *UserRepository) toEntity(model *UserModel) *entity.User {
	return &entity.User{
		ID:                  model.ID,
		Email:               model.Email,
		PasswordHash:        model.PasswordHash,
		Role:                entity.Role(model.Role),
		IsVerified:          model.IsVerified,
		IsActive:            model.IsActive,
		FailedLoginAttempts: model.FailedLoginAttempts,
		LockedUntil:         model.LockedUntil,
		LastLoginAt:         model.LastLoginAt,
		LastLoginIP:         model.LastLoginIP,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
}
