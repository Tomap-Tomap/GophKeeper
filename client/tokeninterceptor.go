package client

import (
	"context"
	"fmt"

	"github.com/Tomap-Tomap/GophKeeper/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const tokenHeader = "authorization"

type tokenInterceptor struct {
	token string
}

func newTokenInterceptor() *tokenInterceptor {
	return &tokenInterceptor{}
}

func (ti *tokenInterceptor) setToken(token string) {
	ti.token = fmt.Sprintf("Bearer %s", token)
}

func (ti *tokenInterceptor) interceptorAddTokenUnary(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if len(ti.token) != 0 {
		ctx = metadata.AppendToOutgoingContext(
			ctx,
			tokenHeader, ti.token,
		)
	}

	err := invoker(ctx, method, req, reply, cc, opts...)

	if err != nil {
		return err
	}

	switch r := reply.(type) {
	case *proto.AuthResponse:
		ti.setToken(r.GetToken())
	case *proto.RegisterResponse:
		ti.setToken(r.GetToken())
	default:
	}

	return err
}

func (ti *tokenInterceptor) interceptorAddTokenStream(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	if len(ti.token) != 0 {
		ctx = metadata.AppendToOutgoingContext(
			ctx,
			tokenHeader, ti.token,
		)
	}

	return streamer(ctx, desc, cc, method, opts...)
}
