package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"auth-service/internal/application/usecase"
	"auth-service/internal/delivery/http/handler"
	"auth-service/internal/delivery/http/middleware"
	"auth-service/internal/delivery/http/router"
	"auth-service/internal/infrastructure/config"
	"auth-service/internal/infrastructure/logger"
	"auth-service/internal/infrastructure/persistence/postgres"
	"auth-service/internal/infrastructure/security"

	"go.uber.org/zap"
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

	authHandler := handler.NewAuthHandler(authUseCase, log, cfg)
	healthHandler := handler.NewHealthHandler()
	authMiddleware := middleware.NewAuthMiddleware(tokenService, tokenBlacklistRepo, userRepo)

	r := router.NewRouter(authHandler, healthHandler, authMiddleware, log, cfg)
	engine := r.Setup()

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Info("starting server", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server", zap.Error(err))
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", zap.Error(err))
	}

	log.Info("server stopped gracefully")
}
