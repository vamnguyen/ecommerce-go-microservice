package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	proto "auth-service/gen/go"
	"auth-service/internal/application/usecase"
	grpcHandler "auth-service/internal/delivery/grpc/handler"
	"auth-service/internal/delivery/grpc/interceptor"
	"auth-service/internal/infrastructure/config"
	"auth-service/internal/infrastructure/logger"
	"auth-service/internal/infrastructure/persistence/postgres"
	"auth-service/internal/infrastructure/security"
	"auth-service/internal/infrastructure/telemetry"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
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
	tokenService, err := security.NewJWTService(
		cfg.JWT.Algorithm,
		cfg.JWT.PrivateKeyPath,
		cfg.JWT.PublicKeyPath,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)
	if err != nil {
		log.Error("failed to initialize JWT service", zap.Error(err))
		panic(err)
	}

	// --- Telemetry Initialization ---
	shutdownTelemetry, err := telemetry.Init("auth-service", cfg.Telemetry.CollectorAddr)
	if err != nil {
		log.Error("failed to initialize telemetry", zap.Error(err))
		// Don't panic, just log error and continue without telemetry if needed, or panic if strict.
		// panic(err)
	}
	defer func() {
		if err := shutdownTelemetry(context.Background()); err != nil {
			log.Error("failed to shutdown telemetry", zap.Error(err))
		}
	}()

	// --- Metrics Server ---
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsPort := ":9090" // Separate port for metrics
		log.Info("starting metrics server", zap.String("port", metricsPort))
		if err := http.ListenAndServe(metricsPort, nil); err != nil {
			log.Error("failed to start metrics server", zap.Error(err))
		}
	}()

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
		grpc.StatsHandler(otelgrpc.NewServerHandler()), // OpenTelemetry StatsHandler
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
