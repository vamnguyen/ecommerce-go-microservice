package usecase

import (
	"context"
	"log"

	"user-service/internal/application/dto"
	"user-service/internal/domain/entity"
	domainErr "user-service/internal/domain/errors"
	"user-service/internal/domain/repository"

	"github.com/google/uuid"
)

type UserUseCase struct {
	profileRepo repository.UserProfileRepository
}

func NewUserUseCase(profileRepo repository.UserProfileRepository) *UserUseCase {
	return &UserUseCase{
		profileRepo: profileRepo,
	}
}

func (uc *UserUseCase) GetProfile(ctx context.Context, userID string) (*dto.UserProfileDTO, error) {
	log.Printf("Getting profile for user ID: %s", userID)
	userUUID, err := uuid.Parse(userID)
	log.Printf("Parsed User UUID: %+v", userUUID)
	if err != nil {
		return nil, domainErr.ErrInvalidInput
	}

	profile, err := uc.profileRepo.FindByUserID(ctx, userUUID)
	if err != nil {
		// If profile doesn't exist, create a default empty profile
		if err == domainErr.ErrProfileNotFound {
			log.Printf("Profile not found, creating default profile for user: %s", userID)
			profile = entity.NewUserProfile(
				userUUID,
				"", // FirstName
				"", // LastName
				"", // Phone
				"", // Address
				"", // City
				"", // Country
				"", // PostalCode
			)
			if err := uc.profileRepo.Create(ctx, profile); err != nil {
				log.Printf("Error creating default profile: %v", err)
				return nil, err
			}
			log.Printf("Default profile created successfully")
		} else {
			log.Printf("Error finding profile: %v", err)
			return nil, err
		}
	}

	return &dto.UserProfileDTO{
		UserID:     profile.UserID.String(),
		FirstName:  profile.FirstName,
		LastName:   profile.LastName,
		Phone:      profile.Phone,
		Address:    profile.Address,
		City:       profile.City,
		Country:    profile.Country,
		PostalCode: profile.PostalCode,
		CreatedAt:  profile.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  profile.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (uc *UserUseCase) UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return domainErr.ErrInvalidInput
	}

	profile, err := uc.profileRepo.FindByUserID(ctx, userUUID)
	if err != nil {
		// If profile doesn't exist, create it
		if err == domainErr.ErrProfileNotFound {
			profile = entity.NewUserProfile(
				userUUID,
				req.FirstName,
				req.LastName,
				req.Phone,
				req.Address,
				req.City,
				req.Country,
				req.PostalCode,
			)
			return uc.profileRepo.Create(ctx, profile)
		}
		return err
	}

	profile.Update(
		req.FirstName,
		req.LastName,
		req.Phone,
		req.Address,
		req.City,
		req.Country,
		req.PostalCode,
	)

	return uc.profileRepo.Update(ctx, profile)
}

func (uc *UserUseCase) GetUser(ctx context.Context, userID string) (*dto.UserProfileDTO, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, domainErr.ErrInvalidInput
	}

	profile, err := uc.profileRepo.FindByUserID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	return &dto.UserProfileDTO{
		UserID:     profile.UserID.String(),
		FirstName:  profile.FirstName,
		LastName:   profile.LastName,
		Phone:      profile.Phone,
		Address:    profile.Address,
		City:       profile.City,
		Country:    profile.Country,
		PostalCode: profile.PostalCode,
		CreatedAt:  profile.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  profile.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (uc *UserUseCase) ListUsers(ctx context.Context, page, pageSize int32) (*dto.ListUsersResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	limit := int(pageSize)
	offset := int((page - 1) * pageSize)

	profiles, total, err := uc.profileRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	users := make([]dto.UserInfoDTO, len(profiles))
	for i, profile := range profiles {
		users[i] = dto.UserInfoDTO{
			UserID:    profile.UserID.String(),
			FirstName: profile.FirstName,
			LastName:  profile.LastName,
			Phone:     profile.Phone,
			CreatedAt: profile.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return &dto.ListUsersResponse{
		Users:    users,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

