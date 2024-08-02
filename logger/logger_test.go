//go:build unit

package logger

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/interop"
	testgrpc "google.golang.org/grpc/interop/grpc_testing"
)

func Test_loggingResponseWriter_Write(t *testing.T) {
	type fields struct {
		ResponseWriter http.ResponseWriter
		bytes          int
		code           int
		wroteHeader    bool
	}
	type args struct {
		b []byte
	}
	type want struct {
		wantRet  int
		wantSize int
		wantErr  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "test zero",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
				code:           200,
				wroteHeader:    false,
			},
			args:    args{[]byte{}},
			want:    want{0, 0, ""},
			wantErr: false,
		},
		{
			name: "test not zero",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
				code:           200,
				wroteHeader:    false,
			},
			args:    args{[]byte{1, 2, 3}},
			want:    want{3, 3, ""},
			wantErr: false,
		},
		{
			name: "test not empty resData",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
				bytes:          3,
				code:           200,
				wroteHeader:    false,
			},
			args:    args{[]byte{1, 2, 3}},
			want:    want{3, 6, ""},
			wantErr: false,
		},
		{
			name: "test not 200 code",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
				bytes:          3,
				code:           400,
				wroteHeader:    true,
			},
			args:    args{[]byte{1, 2, 3}},
			want:    want{3, 6, string([]byte{1, 2, 3})},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &loggingResponseWriter{
				ResponseWriter: tt.fields.ResponseWriter,
				bytes:          tt.fields.bytes,
				code:           tt.fields.code,
				wroteHeader:    tt.fields.wroteHeader,
			}
			got, err := r.Write(tt.args.b)

			if tt.wantErr {
				require.Error(t, err)
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want.wantRet, got)
			assert.Equal(t, tt.want.wantSize, r.bytes)
			assert.Equal(t, tt.want.wantErr, r.error)
		})
	}
}

func Test_loggingResponseWriter_WriteHeader(t *testing.T) {
	type fields struct {
		ResponseWriter http.ResponseWriter
		code           int
	}
	type args struct {
		statusCode int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "test 200",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
			},
			args: args{200},
			want: 200,
		},
		{
			name: "test 400",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
				code:           123,
			},
			args: args{400},
			want: 400,
		},
		{
			name: "test not empty resData",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
			},
			args: args{400},
			want: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &loggingResponseWriter{
				ResponseWriter: tt.fields.ResponseWriter,
				code:           tt.fields.code,
			}
			r.WriteHeader(tt.args.statusCode)

			assert.Equal(t, tt.want, r.code)
		})
	}
}

func TestInitialize(t *testing.T) {
	type args struct {
		level string
	}
	tests := []struct {
		name    string
		path    string
		args    args
		wantErr bool
	}{
		{
			name:    "positive test",
			path:    "stderr",
			args:    args{"INFO"},
			wantErr: false,
		},
		{
			name:    "negative test",
			path:    "stderr",
			args:    args{"TEST"},
			wantErr: true,
		},
		{
			name:    "negative test",
			path:    "errorPath#21231",
			args:    args{"INFO"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Initialize(tt.args.level, tt.path)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type testingSink struct {
	*bytes.Buffer
}

func (s *testingSink) Close() error { return nil }
func (s *testingSink) Sync() error  { return nil }

func TestInterceptorLogger(t *testing.T) {
	sink := &testingSink{new((bytes.Buffer))}
	zap.RegisterSink("testingInceptor", func(u *url.URL) (zap.Sink, error) { return sink, nil })
	Initialize("INFO", "testingInceptor://")

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	s := grpc.NewServer(grpc.UnaryInterceptor(UnaryInterceptorLogger))

	testgrpc.RegisterTestServiceServer(
		s,
		interop.NewTestServer(),
	)

	go func() {
		if err := s.Serve(lis); err != nil {
			require.FailNow(t, err.Error())
		}
	}()

	defer s.Stop()

	t.Run("positive test", func(t *testing.T) {
		conn, err := grpc.NewClient(
			lis.Addr().String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		require.NoError(t, err)
		defer conn.Close()

		client := testgrpc.NewTestServiceClient(conn)
		interop.DoEmptyUnaryCall(context.Background(), client)
		require.NoError(t, err)

		want := []string{
			`"msg":"Got incoming grpc request"`,
			`"msg":"Sending grpc response"`,
		}
		logs := sink.String()

		for _, val := range want {
			assert.Contains(t, logs, val)
		}
	})
}
