// Package logger defines structures and handles for logging.
package logger

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// Log it's singleton variable for working with logs.
var Log *zap.Logger = zap.NewNop()

// Initialize do initialize log variable.
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

type loggingResponseWriter struct {
	http.ResponseWriter
	error       string
	code        int
	bytes       int
	wroteHeader bool
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	if r.code >= 300 {
		r.error = string(b)
	}

	size, err := r.ResponseWriter.Write(b)
	r.bytes += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	if !r.wroteHeader {
		r.code = statusCode
		r.wroteHeader = true
		r.ResponseWriter.WriteHeader(statusCode)
	}
}

// UnaryInterceptorLogger this is an interceptor for logging a request
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
	} else {
		duration := time.Since(start)

		Log.Info("Sending grpc response",
			zap.String("duration", duration.String()),
		)
	}

	return
}

func StreamInterceptorLogger(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	start := time.Now()

	Log.Info("Got incoming stream grpc request",
		zap.String("full method", info.FullMethod),
	)

	err = handler(srv, stream)

	if err != nil {
		Log.Warn("Failed stream request", zap.Error(err))
	} else {
		duration := time.Since(start)

		Log.Info("Sending stream grpc response",
			zap.String("duration", duration.String()),
		)
	}

	return
}
