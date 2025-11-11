package handler

import (
	proto "auth-service/gen/go"
	"auth-service/internal/application/dto"
	"auth-service/internal/application/usecase"
	"auth-service/internal/delivery/grpc/interceptor"
	"context"
)

type GRPCHandler struct {
	proto.UnimplementedAuthServiceServer
	authUsecase usecase.AuthUseCase
}

func NewGRPCHandler(authUsecase usecase.AuthUseCase) *GRPCHandler {
	return &GRPCHandler{
		authUsecase: authUsecase,
	}
}

func (h *GRPCHandler) HealthCheck(ctx context.Context, req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	return &proto.HealthCheckResponse{Status: "healthy", Service: "auth service"}, nil
}

func (h *GRPCHandler) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	registerDTO := dto.RegisterRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	err := h.authUsecase.Register(ctx, registerDTO)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &proto.RegisterResponse{Message: "user registered successfully"}, nil
}

func (h *GRPCHandler) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	loginDTO := dto.LoginRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	ipAddress := interceptor.GetClientIPFromContext(ctx)
	userAgent := interceptor.GetUserAgentFromContext(ctx)

	result, err := h.authUsecase.Login(ctx, loginDTO, ipAddress, userAgent)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &proto.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

func (h *GRPCHandler) RefreshToken(ctx context.Context, req *proto.RefreshTokenRequest) (*proto.RefreshTokenResponse, error) {
	ipAddress := interceptor.GetClientIPFromContext(ctx)
	userAgent := interceptor.GetUserAgentFromContext(ctx)
	accessToken := interceptor.GetAccessTokenFromContext(ctx)

	result, err := h.authUsecase.RefreshToken(ctx, req.GetRefreshToken(), accessToken, ipAddress, userAgent)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &proto.RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

func (h *GRPCHandler) Logout(ctx context.Context, req *proto.LogoutRequest) (*proto.LogoutResponse, error) {
	userID, err := interceptor.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ipAddress := interceptor.GetClientIPFromContext(ctx)
	userAgent := interceptor.GetUserAgentFromContext(ctx)
	accessToken := interceptor.GetAccessTokenFromContext(ctx)

	if err := h.authUsecase.Logout(ctx, userID, req.GetRefreshToken(), accessToken, ipAddress, userAgent); err != nil {
		return nil, toGRPCError(err)
	}

	return &proto.LogoutResponse{}, nil
}

func (h *GRPCHandler) LogoutAll(ctx context.Context, req *proto.LogoutAllRequest) (*proto.LogoutAllResponse, error) {
	userID, err := interceptor.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ipAddress := interceptor.GetClientIPFromContext(ctx)
	userAgent := interceptor.GetUserAgentFromContext(ctx)
	accessToken := interceptor.GetAccessTokenFromContext(ctx)

	if err := h.authUsecase.LogoutAll(ctx, userID, accessToken, ipAddress, userAgent); err != nil {
		return nil, toGRPCError(err)
	}

	return &proto.LogoutAllResponse{}, nil
}

func (h *GRPCHandler) GetMe(ctx context.Context, req *proto.GetMeRequest) (*proto.GetMeResponse, error) {
	userID, err := interceptor.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user, err := h.authUsecase.GetMe(ctx, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &proto.GetMeResponse{
		Id:         user.ID,
		Email:      user.Email,
		Role:       user.Role,
		IsVerified: user.IsVerified,
		IsActive:   user.IsActive,
		CreatedAt:  user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (h *GRPCHandler) ChangePassword(ctx context.Context, req *proto.ChangePasswordRequest) (*proto.ChangePasswordResponse, error) {
	userID, err := interceptor.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	changePasswordDTO := dto.ChangePasswordRequest{
		OldPassword: req.GetOldPassword(),
		NewPassword: req.GetNewPassword(),
	}

	if err := h.authUsecase.ChangePassword(ctx, userID, changePasswordDTO); err != nil {
		return nil, toGRPCError(err)
	}

	return &proto.ChangePasswordResponse{Message: "password changed successfully"}, nil
}

func (h *GRPCHandler) GetPublicKey(ctx context.Context, req *proto.GetPublicKeyRequest) (*proto.GetPublicKeyResponse, error) {
	return &proto.GetPublicKeyResponse{
		PublicKey: h.authUsecase.GetPublicKey(ctx),
		Algorithm: h.authUsecase.GetAlgorithm(),
	}, nil
}
