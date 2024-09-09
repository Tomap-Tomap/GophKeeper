//go:build unit

package logger

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		output    string
		expectErr bool
	}{
		{
			name:      "valid level and output",
			level:     "info",
			output:    "stderr",
			expectErr: false,
		},
		{
			name:      "invalid level",
			level:     "invalid_level",
			output:    "stderr",
			expectErr: true,
		},
		{
			name:      "invalid output path",
			level:     "info",
			output:    "/invalid/path/to/log",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Initialize(tt.level, tt.output)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			core, recorded := observer.New(zap.InfoLevel)
			logger := zap.New(core)

			switch tt.level {
			case "debug":
				logger.Debug("debug message")
				logger.Info("info message")
			case "info":
				logger.Info("info message")
				logger.Warn("warn message")
			case "warn":
				logger.Warn("warn message")
				logger.Error("error message")
			case "error":
				logger.Error("error message")
			}

			if tt.level == "debug" {
				assert.Equal(t, 2, recorded.Len())
			} else if tt.level == "info" {
				assert.Equal(t, 2, recorded.Len())
			} else if tt.level == "warn" {
				assert.Equal(t, 2, recorded.Len())
			} else if tt.level == "error" {
				assert.Equal(t, 1, recorded.Len())
			}
		})
	}
}

type mockProtoMessage struct {
	proto.Message
}

func TestUnaryInterceptorLogger(t *testing.T) {
	tests := []struct {
		name        string
		req         any
		handler     grpc.UnaryHandler
		expectedLog []struct {
			expectedEntry string
			expectedLevel zapcore.Level
		}
		expectedError error
	}{
		{
			name:    "valid proto message",
			req:     &mockProtoMessage{},
			handler: func(ctx context.Context, req any) (any, error) { return "response", nil },
			expectedLog: []struct {
				expectedEntry string
				expectedLevel zapcore.Level
			}{
				{
					expectedEntry: "Got incoming grpc request",
					expectedLevel: zap.InfoLevel,
				},
				{
					expectedEntry: "Sending grpc response",
					expectedLevel: zap.InfoLevel,
				},
			},

			expectedError: nil,
		},
		{
			name: "invalid message type",
			req:  "invalid request",
			handler: func(ctx context.Context, req any) (any, error) {
				return "response", nil
			},
			expectedLog: []struct {
				expectedEntry string
				expectedLevel zapcore.Level
			}{
				{
					expectedEntry: "Payload is not a google.golang.org/protobuf/proto.Message; programmatic error?",
					expectedLevel: zap.WarnLevel,
				},
				{
					expectedEntry: "Sending grpc response",
					expectedLevel: zap.InfoLevel,
				},
			},
			expectedError: nil,
		},
		{
			name: "handler returns an error",
			req:  &mockProtoMessage{},
			handler: func(ctx context.Context, req any) (any, error) {
				return nil, status.Errorf(codes.Internal, "original error")
			},
			expectedLog: []struct {
				expectedEntry string
				expectedLevel zapcore.Level
			}{
				{
					expectedEntry: "Got incoming grpc request",
					expectedLevel: zap.InfoLevel,
				},
				{
					expectedEntry: "Failed request",
					expectedLevel: zap.WarnLevel,
				},
			},
			expectedError: status.Errorf(codes.Internal, "internal error on method %s", "/some/method"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка наблюдателя для логов
			core, observedLogs := observer.New(zap.InfoLevel)
			Log = zap.New(core)

			// Создание метода дя информации о методе grpc, здесь это фиктивные данные
			info := &grpc.UnaryServerInfo{
				FullMethod: "/some/method",
			}

			// Вызов целевой функции
			_, err := UnaryInterceptorLogger(context.Background(), tt.req, info, tt.handler)

			// Проверка возвращенной ошибки (если ожидается)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			// Проверка логов
			logs := observedLogs.All()
			assert.True(t, len(logs) > 0)

			for i := range logs {
				assert.Equal(t, tt.expectedLog[i].expectedEntry, logs[i].Message)
				assert.Equal(t, tt.expectedLog[i].expectedLevel, logs[i].Level)
			}
		})
	}
}

type mockServerStream struct {
	grpc.ServerStream
}

func (m *mockServerStream) SetHeader(md metadata.MD) error {
	return nil
}

func (m *mockServerStream) SendHeader(md metadata.MD) error {
	return nil
}

func (m *mockServerStream) SetTrailer(md metadata.MD) {}

func TestStreamInterceptorLogger(t *testing.T) {
	logger := zaptest.NewLogger(t)
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	tests := []struct {
		name          string
		handler       grpc.StreamHandler
		expectedError error
		expectedCode  codes.Code
	}{
		{
			name: "successful stream",
			handler: func(srv interface{}, stream grpc.ServerStream) error {
				return nil
			},
			expectedError: nil,
			expectedCode:  codes.OK,
		},
		{
			name: "failed stream with internal error",
			handler: func(srv interface{}, stream grpc.ServerStream) error {
				return status.Error(codes.Internal, "internal error")
			},
			expectedError: status.Error(codes.Internal, "internal error on method /test.TestService/TestMethod"),
			expectedCode:  codes.Internal,
		},
		{
			name: "failed stream with other error",
			handler: func(srv interface{}, stream grpc.ServerStream) error {
				return status.Error(codes.InvalidArgument, "invalid argument")
			},
			expectedError: status.Error(codes.InvalidArgument, "invalid argument"),
			expectedCode:  codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &grpc.StreamServerInfo{
				FullMethod: "/test.TestService/TestMethod",
			}
			srv := struct{}{}
			stream := &mockServerStream{}

			start := time.Now()
			err := StreamInterceptorLogger(srv, stream, info, tt.handler)

			assert.Equal(t, tt.expectedError, err)
			if err != nil {
				assert.Equal(t, tt.expectedCode, status.Code(err))
			}
			duration := time.Since(start)
			t.Logf("Duration: %v", duration)
		})
	}
}
