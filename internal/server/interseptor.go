package server

import (
	"context"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Claims struct {
	Email string `json:"sub"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// AuthInterceptor returns a unary interceptor that validates JWT tokens.
func AuthInterceptor(secret string, unprotected map[string]bool, adminOnly map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if unprotected[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		raw := strings.TrimPrefix(authHeader[0], "Bearer ")
		token, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		claims, ok := token.Claims.(*Claims)
		if !ok || !token.Valid {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		if adminOnly[info.FullMethod] && claims.Role != "admin" {
			return nil, status.Error(codes.PermissionDenied, "admin role required")
		}

		return handler(ctx, req)
	}
}
