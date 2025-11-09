package handler

import (
	"auth-service/gen/go/proto"
	"auth-service/internal/application/dto"
	"auth-service/internal/application/usecase"
	"context"
)

// GRPCHandler triển khai interface AuthServiceServer được sinh ra bởi protoc.
type GRPCHandler struct {
	proto.UnimplementedAuthServiceServer // Bắt buộc phải có để tương thích về phía trước
	authUsecase                          usecase.AuthUseCase
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

	result, err := h.authUsecase.Login(ctx, loginDTO, req.GetIpAddress(), req.GetUserAgent())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &proto.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}
