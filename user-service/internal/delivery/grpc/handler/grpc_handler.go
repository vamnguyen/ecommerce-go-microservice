package handler

import (
	"context"

	proto "user-service/gen/go"
	"user-service/internal/application/dto"
	"user-service/internal/application/usecase"
	"user-service/internal/delivery/grpc/interceptor"
)

type GRPCHandler struct {
	proto.UnimplementedUserServiceServer
	userUsecase usecase.UserUseCase
}

func NewGRPCHandler(userUsecase usecase.UserUseCase) *GRPCHandler {
	return &GRPCHandler{
		userUsecase: userUsecase,
	}
}

func (h *GRPCHandler) HealthCheck(ctx context.Context, req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	return &proto.HealthCheckResponse{Status: "healthy", Service: "user service"}, nil
}

func (h *GRPCHandler) GetProfile(ctx context.Context, req *proto.GetProfileRequest) (*proto.GetProfileResponse, error) {
	userID, err := interceptor.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	profile, err := h.userUsecase.GetProfile(ctx, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	email := interceptor.GetUserEmailFromContext(ctx)

	return &proto.GetProfileResponse{
		UserId:     profile.UserID,
		Email:      email,
		FirstName:  profile.FirstName,
		LastName:   profile.LastName,
		Phone:      profile.Phone,
		Address:    profile.Address,
		City:       profile.City,
		Country:    profile.Country,
		PostalCode: profile.PostalCode,
		CreatedAt:  profile.CreatedAt,
		UpdatedAt:  profile.UpdatedAt,
	}, nil
}

func (h *GRPCHandler) UpdateProfile(ctx context.Context, req *proto.UpdateProfileRequest) (*proto.UpdateProfileResponse, error) {
	userID, err := interceptor.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	updateDTO := dto.UpdateProfileRequest{
		FirstName:  req.GetFirstName(),
		LastName:   req.GetLastName(),
		Phone:      req.GetPhone(),
		Address:    req.GetAddress(),
		City:       req.GetCity(),
		Country:    req.GetCountry(),
		PostalCode: req.GetPostalCode(),
	}

	if err := h.userUsecase.UpdateProfile(ctx, userID, updateDTO); err != nil {
		return nil, toGRPCError(err)
	}

	return &proto.UpdateProfileResponse{Message: "profile updated successfully"}, nil
}

func (h *GRPCHandler) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	profile, err := h.userUsecase.GetUser(ctx, req.GetUserId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &proto.GetUserResponse{
		UserId:     profile.UserID,
		FirstName:  profile.FirstName,
		LastName:   profile.LastName,
		Phone:      profile.Phone,
		Address:    profile.Address,
		City:       profile.City,
		Country:    profile.Country,
		PostalCode: profile.PostalCode,
		CreatedAt:  profile.CreatedAt,
		UpdatedAt:  profile.UpdatedAt,
	}, nil
}

func (h *GRPCHandler) ListUsers(ctx context.Context, req *proto.ListUsersRequest) (*proto.ListUsersResponse, error) {
	page := req.GetPage()
	if page < 1 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize < 1 {
		pageSize = 10
	}

	result, err := h.userUsecase.ListUsers(ctx, page, pageSize)
	if err != nil {
		return nil, toGRPCError(err)
	}

	users := make([]*proto.UserInfo, len(result.Users))
	for i, user := range result.Users {
		users[i] = &proto.UserInfo{
			UserId:    user.UserID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
			CreatedAt: user.CreatedAt,
		}
	}

	return &proto.ListUsersResponse{
		Users:    users,
		Total:    int32(result.Total),
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}

