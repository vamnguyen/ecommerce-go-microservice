package interceptor

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	UserIDKey      contextKey = "user_id"
	UserEmailKey   contextKey = "user_email"
	UserRoleKey    contextKey = "user_role"
	ClientIPKey    contextKey = "client_ip"
	UserAgentKey   contextKey = "user_agent"
	AccessTokenKey contextKey = "access_token"
)

var publicMethods = map[string]bool{
	"/proto.OrderService/HealthCheck": true,
}

func NewAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		// Extract client info
		if ips := md.Get("x-forwarded-for"); len(ips) > 0 {
			ctx = context.WithValue(ctx, ClientIPKey, ips[0])
		}
		if agents := md.Get("user-agent"); len(agents) > 0 {
			ctx = context.WithValue(ctx, UserAgentKey, agents[0])
		}

		// Public methods don't need authentication
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Check if request was authenticated by Kong Gateway
		if kongConsumerID := md.Get("x-consumer-id"); len(kongConsumerID) > 0 {
			log.Println("✅ Request authenticated by Kong Gateway, consumer:", kongConsumerID[0])

			// Extract JWT token and parse claims (Kong already validated it)
			authHeaders := md.Get("authorization")
			if len(authHeaders) > 0 {
				parts := strings.SplitN(authHeaders[0], " ", 2)
				if len(parts) == 2 {
					token := parts[1]
					// For order-service, we'll extract user info from Kong headers
					if customID := md.Get("x-consumer-custom-id"); len(customID) > 0 {
						ctx = context.WithValue(ctx, UserIDKey, customID[0])
					}
					if username := md.Get("x-consumer-username"); len(username) > 0 {
						ctx = context.WithValue(ctx, UserEmailKey, username[0])
					}
					ctx = context.WithValue(ctx, AccessTokenKey, token)
					return handler(ctx, req)
				}
			}

			return nil, status.Error(codes.Internal, "failed to extract user info from JWT")
		}

		// If no Kong consumer header, request is unauthenticated
		log.Println("❌ Request not authenticated by Kong Gateway")
		return nil, status.Error(codes.Unauthenticated, "unauthorized: request must go through API gateway")
	}
}

func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return "", status.Error(codes.Unauthenticated, "user id not found in context")
	}
	return userID, nil
}

func GetClientIPFromContext(ctx context.Context) string {
	ip, _ := ctx.Value(ClientIPKey).(string)
	return ip
}

func GetUserAgentFromContext(ctx context.Context) string {
	agent, _ := ctx.Value(UserAgentKey).(string)
	return agent
}

func GetAccessTokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(AccessTokenKey).(string)
	return token
}

