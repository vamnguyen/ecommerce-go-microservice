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
	"/proto.AuthService/HealthCheck":  true,
	"/proto.AuthService/Register":     true,
	"/proto.AuthService/Login":        true,
	"/proto.AuthService/GetPublicKey": true,
}

func NewAuthInterceptor(tokenService TokenValidator) grpc.UnaryServerInterceptor {
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
					// Parse claims without validation (Kong already did it)
					claims, err := tokenService.ExtractClaims(token)
					if err == nil {
						ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
						ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
						ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
						ctx = context.WithValue(ctx, AccessTokenKey, token)
						log.Println("📋 Extracted user info - UserID:", claims.UserID, "Email:", claims.Email,
							"Role:", claims.Role)
						return handler(ctx, req)
					}
					log.Println("⚠️  Failed to extract claims from JWT:", err)
				}
			}

			return nil, status.Error(codes.Internal, "failed to extract user info from JWT")
		}

		// If no Kong consumer header, request is unauthenticated
		log.Println("❌ Request not authenticated by Kong Gateway")
		return nil, status.Error(codes.Unauthenticated, "unauthorized: request must go through API gateway")
	}
}

type TokenValidator interface {
	ExtractClaims(token string) (*TokenClaims, error)
}

type TokenClaims struct {
	UserID string
	Email  string
	Role   string
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
