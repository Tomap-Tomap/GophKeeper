// Package logger provides structures and functionality for application logging.
package logger

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Log is a singleton variable that allows for centralized logging across the application.
var Log *zap.Logger = zap.NewNop()

// Initialize sets up the logging environment by configuring the log level and output path.
// Returns an error if there is an issue with parsing the log level or building the logger
func Initialize(level string, outputPath string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return fmt.Errorf("parse level %s: %w", level, err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	cfg.OutputPaths = []string{outputPath}
	zl, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("build logger: %w", err)
	}

	Log = zl

	return nil
}

// UnaryInterceptorLogger is a gRPC interceptor for logging unary requests and responses.
// This function logs the incoming request, processes it, and then logs the response along with the duration of the request.
func UnaryInterceptorLogger(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	start := time.Now()

	if v, ok := req.(proto.Message); ok {
		Log.Info("Got incoming grpc request",
			zap.String("full method", info.FullMethod),
			zap.Any("body", v),
		)
	} else {
		Log.Warn("Payload is not a google.golang.org/protobuf/proto.Message; programmatic error?",
			zap.String("full method", info.FullMethod))
	}

	resp, err = handler(ctx, req)

	if err != nil {
		Log.Warn("Failed request", zap.Error(err))

		if status.Code(err) == codes.Internal {
			err = status.Errorf(codes.Internal, "internal error on method %s", info.FullMethod)
		}
	} else {
		duration := time.Since(start)

		Log.Info("Sending grpc response",
			zap.String("duration", duration.String()),
		)
	}

	return
}

// StreamInterceptorLogger is a gRPC interceptor for logging streaming requests and responses.
// This function logs the incoming stream request, processes it, and then logs the stream response along with the duration of the request.
func StreamInterceptorLogger(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	start := time.Now()

	Log.Info("Got incoming stream grpc request",
		zap.String("full method", info.FullMethod),
	)

	err = handler(srv, stream)

	if err != nil {
		Log.Warn("Failed stream request", zap.Error(err))

		if status.Code(err) == codes.Internal {
			err = status.Errorf(codes.Internal, "internal error on method %s", info.FullMethod)
		}
	} else {
		duration := time.Since(start)

		Log.Info("Sending stream grpc response",
			zap.String("duration", duration.String()),
		)
	}

	return
}
