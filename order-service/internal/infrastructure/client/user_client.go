package client

import (
	"context"

	"order-service/internal/infrastructure/config"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type UserClient struct {
	conn   *grpc.ClientConn
	health grpc_health_v1.HealthClient
}

func NewUserClient(cfg *config.ServicesConfig) (*UserClient, error) {
	conn, err := grpc.NewClient(
		cfg.UserServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, err
	}

	return &UserClient{
		conn:   conn,
		health: grpc_health_v1.NewHealthClient(conn),
	}, nil
}

// GetUser validates that user-service is available
// For MVP, we just check if the service is healthy
func (c *UserClient) GetUser(ctx context.Context, userID string) error {
	_, err := c.health.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: "user-service",
	})
	return err
}

func (c *UserClient) Close() error {
	return c.conn.Close()
}

