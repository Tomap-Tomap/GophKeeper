//go:build unit

// Package tokener определяет методы и структуры для генерации JWT токенов
package tokener

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/interop"
	testgrpc "google.golang.org/grpc/interop/grpc_testing"
	"google.golang.org/grpc/metadata"
)

func RunGPRCTestServer(t *testing.T, tokener Tokener) (addr string, stopFunc func()) {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	s := grpc.NewServer(grpc.UnaryInterceptor(tokener.UnaryServerInterceptor),
		grpc.StreamInterceptor(tokener.StreamServerInterceptor))

	testgrpc.RegisterTestServiceServer(
		s,
		interop.NewTestServer(),
	)

	go func() {
		if err := s.Serve(lis); err != nil && err != http.ErrServerClosed {
			require.FailNow(t, err.Error())
		}
	}()

	addr = lis.Addr().String()
	stopFunc = s.Stop

	return
}

func CreateGRPCTestClient(t *testing.T, addr string) (client testgrpc.TestServiceClient, closeFunc func() error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client = testgrpc.NewTestServiceClient(conn)
	closeFunc = conn.Close

	return
}

func TestTokener_GetToken(t *testing.T) {
	tokener := NewTokener([]byte("secret"), time.Duration(1))

	token, err := tokener.GetToken("sub")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestTokener_authByGRPCContext(t *testing.T) {
	tokener := NewTokener([]byte("secret"), time.Duration(1*time.Hour))

	t.Run("mssing md error", func(t *testing.T) {
		ctx := context.Background()
		gotCTX, err := tokener.authByGRPCContext(ctx)
		assert.ErrorContains(t, err, "missing metadata")
		assert.Equal(t, ctx, gotCTX)
	})

	t.Run("mssing header error", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{})
		gotCTX, err := tokener.authByGRPCContext(ctx)
		assert.ErrorContains(t, err, "missing authorization")
		assert.Equal(t, ctx, gotCTX)
	})

	t.Run("bad auth string error", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "error"))
		gotCTX, err := tokener.authByGRPCContext(ctx)
		assert.ErrorContains(t, err, "bad authorization string")
		assert.Equal(t, ctx, gotCTX)
	})

	t.Run("request unauthenticated error", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "error error"))
		gotCTX, err := tokener.authByGRPCContext(ctx)
		assert.ErrorContains(t, err, "request unauthenticated with Bearer")
		assert.Equal(t, ctx, gotCTX)
	})

	t.Run("invalid auth token error", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer error"))
		gotCTX, err := tokener.authByGRPCContext(ctx)
		assert.ErrorContains(t, err, "invalid auth token")
		assert.Equal(t, ctx, gotCTX)
	})

	t.Run("positive test", func(t *testing.T) {
		sub := "test"
		token, err := tokener.GetToken(sub)
		require.NoError(t, err)
		tokenHead := fmt.Sprintf("Bearer %s", token)
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", tokenHead))
		gotCTX, err := tokener.authByGRPCContext(ctx)
		require.NoError(t, err)
		md, ok := metadata.FromIncomingContext(gotCTX)
		require.True(t, ok)
		gotSub := md.Get("user_id")
		require.Len(t, gotSub, 1)
		gotSubStr := gotSub[0]
		assert.Equal(t, sub, gotSubStr)
	})
}

func TestTokener_getSubFromToken(t *testing.T) {
	tokener := NewTokener([]byte("secret"), time.Duration(1*time.Hour))

	t.Run("unexpected signing method error", func(t *testing.T) {
		token := jwt.NewWithClaims(
			jwt.SigningMethodNone,
			jwt.RegisteredClaims{
				Subject:   "test",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(1 * time.Hour))),
			},
		)
		tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		sub, err := tokener.getSubFromToken(tokenString)
		assert.Error(t, err)
		assert.Empty(t, sub)
	})

	t.Run("positive test", func(t *testing.T) {
		sub := "test"
		token, err := tokener.GetToken(sub)
		require.NoError(t, err)

		subGot, err := tokener.getSubFromToken(token)
		assert.NoError(t, err)
		assert.Equal(t, sub, subGot)
	})
}

func TestTokener_UnaryServerInterceptor(t *testing.T) {
	tokener := NewTokener([]byte("secret"), time.Duration(1*time.Hour))

	addr, stopServer := RunGPRCTestServer(t, *tokener)
	defer stopServer()

	client, closeClient := CreateGRPCTestClient(t, addr)
	defer closeClient()

	t.Run("auth error", func(t *testing.T) {
		_, err := client.EmptyCall(context.Background(), &testgrpc.Empty{})
		assert.Error(t, err)
	})

	t.Run("positive test", func(t *testing.T) {
		token, err := tokener.GetToken("test")
		require.NoError(t, err)
		tokenHead := fmt.Sprintf("Bearer %s", token)
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", tokenHead))
		_, err = client.EmptyCall(ctx, &testgrpc.Empty{})
		assert.NoError(t, err)
	})
}

func TestTokener_StreamServerInterceptor(t *testing.T) {
	tokener := NewTokener([]byte("secret"), time.Duration(1*time.Hour))

	addr, stopServer := RunGPRCTestServer(t, *tokener)
	defer stopServer()

	client, closeClient := CreateGRPCTestClient(t, addr)
	defer closeClient()

	t.Run("auth error", func(t *testing.T) {
		stream, err := client.StreamingInputCall(context.Background())
		assert.NoError(t, err)
		err = stream.SendMsg(&testgrpc.Empty{})
		assert.NoError(t, err)
		_, err = stream.CloseAndRecv()
		assert.Error(t, err)
	})

	t.Run("auth error", func(t *testing.T) {
		token, err := tokener.GetToken("test")
		require.NoError(t, err)
		tokenHead := fmt.Sprintf("Bearer %s", token)
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", tokenHead))
		stream, err := client.StreamingInputCall(ctx)
		assert.NoError(t, err)
		err = stream.SendMsg(&testgrpc.Empty{})
		assert.NoError(t, err)
		_, err = stream.CloseAndRecv()
		assert.NoError(t, err)
	})
}
