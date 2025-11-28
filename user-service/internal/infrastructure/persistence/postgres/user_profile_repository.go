package postgres

import (
	"context"
	"errors"
	"time"

	"user-service/internal/domain/entity"
	domainErr "user-service/internal/domain/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserProfileModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	FirstName  string
	LastName   string
	Phone      string
	Address    string
	City       string
	Country    string
	PostalCode string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (UserProfileModel) TableName() string {
	return "user_profiles"
}

type UserProfileRepository struct {
	db *gorm.DB
}

func NewUserProfileRepository(db *gorm.DB) *UserProfileRepository {
	return &UserProfileRepository{db: db}
}

func (r *UserProfileRepository) Create(ctx context.Context, profile *entity.UserProfile) error {
	model := r.toModel(profile)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *UserProfileRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserProfile, error) {
	var model UserProfileModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErr.ErrProfileNotFound
		}
		return nil, domainErr.ErrDatabase
	}
	return r.toEntity(&model), nil
}

func (r *UserProfileRepository) Update(ctx context.Context, profile *entity.UserProfile) error {
	model := r.toModel(profile)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return domainErr.ErrDatabase
	}
	return nil
}

func (r *UserProfileRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.UserProfile, error) {
	var model UserProfileModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErr.ErrProfileNotFound
		}
		return nil, domainErr.ErrDatabase
	}
	return r.toEntity(&model), nil
}

func (r *UserProfileRepository) List(ctx context.Context, limit, offset int) ([]*entity.UserProfile, int64, error) {
	var models []UserProfileModel
	var total int64

	if err := r.db.WithContext(ctx).Model(&UserProfileModel{}).Count(&total).Error; err != nil {
		return nil, 0, domainErr.ErrDatabase
	}

	if err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, 0, domainErr.ErrDatabase
	}

	profiles := make([]*entity.UserProfile, len(models))
	for i, model := range models {
		profiles[i] = r.toEntity(&model)
	}

	return profiles, total, nil
}

func (r *UserProfileRepository) toModel(profile *entity.UserProfile) *UserProfileModel {
	return &UserProfileModel{
		ID:         profile.ID,
		UserID:     profile.UserID,
		FirstName:  profile.FirstName,
		LastName:   profile.LastName,
		Phone:      profile.Phone,
		Address:    profile.Address,
		City:       profile.City,
		Country:    profile.Country,
		PostalCode: profile.PostalCode,
		CreatedAt:  profile.CreatedAt,
		UpdatedAt:  profile.UpdatedAt,
	}
}

func (r *UserProfileRepository) toEntity(model *UserProfileModel) *entity.UserProfile {
	return &entity.UserProfile{
		ID:         model.ID,
		UserID:     model.UserID,
		FirstName:  model.FirstName,
		LastName:   model.LastName,
		Phone:      model.Phone,
		Address:    model.Address,
		City:       model.City,
		Country:    model.Country,
		PostalCode: model.PostalCode,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}
}

