package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"auth-service/gen/go/proto"
	"auth-service/internal/application/usecase"
	grpcHandler "auth-service/internal/delivery/grpc/handler"
	"auth-service/internal/delivery/http/handler"
	"auth-service/internal/delivery/http/middleware"
	"auth-service/internal/delivery/http/router"
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

	// Khởi tạo HTTP server
	authHandler := handler.NewAuthHandler(authUseCase, log, cfg)
	healthHandler := handler.NewHealthHandler()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, tokenBlacklistRepo, userRepo)

	r := router.NewRouter(authHandler, healthHandler, authMiddleware, log, cfg)
	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r.Setup(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Khởi tạo gRPC server
	grpcHandler := grpcHandler.NewGRPCHandler(*authUseCase)
	grpcServer := grpc.NewServer()
	proto.RegisterAuthServiceServer(grpcServer, grpcHandler)

	// Chạy cả hai server đồng thời
	// go func() {
	// 	log.Info("starting HTTP server", zap.String("port", cfg.Server.Port))
	// 	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 		log.Error("failed to start HTTP server", zap.Error(err))
	// 		panic(err)
	// 	}
	// }()

	go func() {
		grpcListener, err := net.Listen("tcp", ":"+cfg.Server.GRPCPort)
		if err != nil {
			log.Error("failed to listen for gRPC", zap.Error(err))
			panic(err)
		}
		log.Info("starting gRPC server", zap.String("port", cfg.Server.GRPCPort))
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Error("failed to start gRPC server", zap.Error(err))
			panic(err)
		}
	}()

	// Xử lý Graceful Shutdown cho cả hai server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Sử dụng WaitGroup để chờ cả hai server shutdown xong
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		log.Info("shutting down gRPC server...")
		grpcServer.GracefulStop()
		log.Info("gRPC server stopped")
	}()

	go func() {
		defer wg.Done()
		log.Info("shutting down HTTP server...")
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Error("HTTP server forced to shutdown", zap.Error(err))
		}
		log.Info("HTTP server stopped")
	}()

	wg.Wait()
	log.Info("all servers stopped gracefully")
}
