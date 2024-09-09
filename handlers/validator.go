package handlers

import (
	"context"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Validator struct that embeds the protovalidate.Validator
type Validator struct {
	*protovalidate.Validator
}

// NewValidator creates a new Validator instance with the provided protovalidate.Validator
func NewValidator(validator *protovalidate.Validator) *Validator {
	return &Validator{
		validator,
	}
}

// UnaryServerInterceptor is a method of Validator that acts as a gRPC Unary Server Interceptor
func (v *Validator) UnaryServerInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	msg, ok := req.(protoreflect.ProtoMessage)

	if ok {
		err := v.Validate(msg)

		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return handler(ctx, req)
}
