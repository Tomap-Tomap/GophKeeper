// Package tokener определяет методы и структуры для генерации JWT токенов
package tokener

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tomap-Tomap/GophKeeper/proto"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	headerAuthorize    = "authorization"
	expectedAuthScheme = "Bearer"
)

// Tokener структура для работы с JWT токенами
type Tokener struct {
	secret string
	exp    time.Duration
}

// NewTokener аллоцирует новый Tokener
func NewTokener(secret string, exp time.Duration) *Tokener {
	return &Tokener{
		secret: secret,
		exp:    exp,
	}
}

// GetToken генирирует новый токен для sub
func (t *Tokener) GetToken(sub string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Subject:   sub,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.exp)),
		},
	)
	tokenString, err := token.SignedString([]byte(t.secret))

	if err != nil {
		return "", fmt.Errorf("cannot sign token: %w", err)
	}

	return tokenString, nil
}

// StreamServerInterceptor перехватчик для стрима сервера grpc
func (t *Tokener) StreamServerInterceptor(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := stream.Context()
	newCtx, err := t.authByGRPCContext(ctx)

	if err != nil {
		return err
	}

	wrapped := proto.WrapServerStream(stream)
	wrapped.WrappedContext = newCtx

	return handler(srv, wrapped)
}

// UnaryServerInterceptor перехватчик для унарного сервера grpc
func (t *Tokener) UnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if strings.Contains(info.FullMethod, "Register") || strings.Contains(info.FullMethod, "Auth") {
		return handler(ctx, req)
	}

	newCtx, err := t.authByGRPCContext(ctx)

	if err != nil {
		return nil, err
	}

	return handler(newCtx, req)
}

func (t *Tokener) getSubFromToken(token string) (string, bool) {
	claims := &jwt.RegisteredClaims{}
	jwtToken, err := jwt.ParseWithClaims(token, claims, func(jwtT *jwt.Token) (interface{}, error) {
		if _, ok := jwtT.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", jwtT.Header["alg"])
		}

		return []byte(t.secret), nil
	})

	if err != nil {
		return "", false
	}

	if !jwtToken.Valid {
		return "", false
	}

	return claims.Subject, true
}

func (t *Tokener) authByGRPCContext(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return ctx, status.Error(codes.Unauthenticated, "missing metadata")
	}

	auth := md.Get(headerAuthorize)

	if len(auth) == 0 {
		return ctx, status.Errorf(codes.Unauthenticated, "missing %s", headerAuthorize)
	}

	scheme, token, found := strings.Cut(auth[0], " ")

	if !found {
		return ctx, status.Error(codes.Unauthenticated, "bad authorization string")
	}
	if !strings.EqualFold(scheme, expectedAuthScheme) {
		return ctx, status.Errorf(codes.Unauthenticated, "request unauthenticated with %s", expectedAuthScheme)
	}

	sub, tokenValid := t.getSubFromToken(token)

	if !tokenValid {
		return ctx, status.Error(codes.PermissionDenied, "invalid auth token")
	}

	newCtx := metadata.NewIncomingContext(ctx, metadata.Pairs("user_id", sub))

	return newCtx, nil
}
