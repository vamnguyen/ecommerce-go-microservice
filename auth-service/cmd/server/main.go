package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"auth-service/gen/go/proto"
	"auth-service/internal/application/usecase"
	grpcHandler "auth-service/internal/delivery/grpc/handler"
	"auth-service/internal/delivery/grpc/interceptor"
	"auth-service/internal/infrastructure/config"
	"auth-service/internal/infrastructure/logger"
	"auth-service/internal/infrastructure/persistence/postgres"
	"auth-service/internal/infrastructure/security"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	log, err := logger.New(cfg.Environment)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer log.Close()

	db, err := postgres.NewDatabase(&cfg.Database)
	if err != nil {
		log.Error("failed to connect to database", zap.Error(err))
		panic(err)
	}
	if err := postgres.AutoMigrate(db); err != nil {
		log.Error("failed to run migrations", zap.Error(err))
		panic(err)
	}

	userRepo := postgres.NewUserRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)
	tokenBlacklistRepo := postgres.NewTokenBlacklistRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)

	passwordService := security.NewBcryptPasswordService()
	tokenService := security.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	authUseCase := usecase.NewAuthUseCase(
		userRepo,
		refreshTokenRepo,
		tokenBlacklistRepo,
		auditLogRepo,
		passwordService,
		tokenService,
		usecase.AuthConfig{
			MaxLoginAttempts:    cfg.Security.MaxLoginAttempts,
			AccountLockDuration: cfg.Security.AccountLockDuration,
		},
	)

	grpcHandler := grpcHandler.NewGRPCHandler(*authUseCase)
	
	tokenValidator := interceptor.NewTokenServiceAdapter(tokenService)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.NewAuthInterceptor(tokenValidator)),
	)
	proto.RegisterAuthServiceServer(grpcServer, grpcHandler)

	grpcListener, err := net.Listen("tcp", ":"+cfg.Server.GRPCPort)
	if err != nil {
		log.Error("failed to listen for gRPC", zap.Error(err))
		panic(err)
	}

	go func() {
		log.Info("starting gRPC server", zap.String("port", cfg.Server.GRPCPort))
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Error("failed to start gRPC server", zap.Error(err))
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down gRPC server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Info("gRPC server stopped gracefully")
	case <-ctx.Done():
		log.Warn("gRPC server forced to shutdown")
	}

	log.Info("server stopped")
}
