package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	proto "order-service/gen/go"
	"order-service/internal/application/usecase"
	grpcHandler "order-service/internal/delivery/grpc/handler"
	"order-service/internal/delivery/grpc/interceptor"
	"order-service/internal/infrastructure/client"
	"order-service/internal/infrastructure/config"
	"order-service/internal/infrastructure/logger"
	"order-service/internal/infrastructure/persistence/postgres"
	"order-service/internal/infrastructure/telemetry"

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

	orderRepo := postgres.NewOrderRepository(db)

	// Initialize user-service client
	userClient, err := client.NewUserClient(&cfg.Services)
	if err != nil {
		log.Error("failed to initialize user client", zap.Error(err))
		panic(err)
	}
	defer userClient.Close()

	// --- Telemetry Initialization ---
	shutdownTelemetry, err := telemetry.Init("order-service", cfg.Telemetry.CollectorAddr)
	if err != nil {
		log.Error("failed to initialize telemetry", zap.Error(err))
	}
	defer func() {
		if err := shutdownTelemetry(context.Background()); err != nil {
			log.Error("failed to shutdown telemetry", zap.Error(err))
		}
	}()

	// --- Metrics Server ---
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsPort := ":9090"
		log.Info("starting metrics server", zap.String("port", metricsPort))
		if err := http.ListenAndServe(metricsPort, nil); err != nil {
			log.Error("failed to start metrics server", zap.Error(err))
		}
	}()

	orderUseCase := usecase.NewOrderUseCase(orderRepo, userClient)

	grpcHandler := grpcHandler.NewGRPCHandler(*orderUseCase)

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(interceptor.NewAuthInterceptor()),
	)
	proto.RegisterOrderServiceServer(grpcServer, grpcHandler)

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

