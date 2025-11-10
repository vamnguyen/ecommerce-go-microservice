package interceptor

import (
	"context"
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
	"/proto.AuthService/HealthCheck": true,
	"/proto.AuthService/Register":    true,
	"/proto.AuthService/Login":       true,
}

func NewAuthInterceptor(tokenService TokenValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		if ips := md.Get("x-forwarded-for"); len(ips) > 0 {
			ctx = context.WithValue(ctx, ClientIPKey, ips[0])
		}

		if agents := md.Get("user-agent"); len(agents) > 0 {
			ctx = context.WithValue(ctx, UserAgentKey, agents[0])
		}

		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		authHeader := authHeaders[0]
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		token := parts[1]
		claims, err := tokenService.ValidateAccessToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
		ctx = context.WithValue(ctx, AccessTokenKey, token)

		return handler(ctx, req)
	}
}

type TokenValidator interface {
	ValidateAccessToken(token string) (*TokenClaims, error)
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
