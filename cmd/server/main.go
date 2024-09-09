package main

import (
	"context"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tomap-Tomap/GophKeeper/handlers"
	"github.com/Tomap-Tomap/GophKeeper/hasher"
	"github.com/Tomap-Tomap/GophKeeper/logger"
	"github.com/Tomap-Tomap/GophKeeper/parameters"
	proto "github.com/Tomap-Tomap/GophKeeper/proto/gophkeeper/v1"
	"github.com/Tomap-Tomap/GophKeeper/storage"
	"github.com/Tomap-Tomap/GophKeeper/tokener"
	"github.com/bufbuild/protovalidate-go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	_ "google.golang.org/grpc/encoding/gzip"
)

func main() {
	if err := logger.Initialize("INFO", "stderr"); err != nil {
		panic(err)
	}

	p := parameters.ParseFlagsServer()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	eg, egCtx := errgroup.WithContext(ctx)

	s, err := storage.NewStorage(egCtx, p.DSN)

	if err != nil {
		logger.Log.Fatal("Cannot create new DB storage", zap.Error(err))
	}

	h := hasher.NewHasher()
	t := tokener.NewTokener([]byte(p.TokenSecret), time.Duration(time.Duration(p.TokenDuration)*time.Minute))

	fs := storage.NewFileStorage(p.PathToFileStorage, int(p.ChunkSize))

	listen, err := net.Listen("tcp", p.GRPCAddr)
	if err != nil {
		logger.Log.Fatal("Cannot create listener", zap.Error(err))
	}

	validator, err := protovalidate.New()

	if err != nil {
		logger.Log.Fatal("Cannot create validator", zap.Error(err))
	}

	v := handlers.NewValidator(validator)

	gs := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logger.UnaryInterceptorLogger,
			t.UnaryServerInterceptor,
			v.UnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			logger.StreamInterceptorLogger,
			t.StreamServerInterceptor,
		),
	)

	proto.RegisterGophKeeperServiceServer(gs, handlers.NewGophKeeperHandler(s, h, t, fs, *storage.NewRetryPolicy(3, 5, 3), 75))

	eg.Go(func() error {
		err := gs.Serve(listen)
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		<-egCtx.Done()
		gs.GracefulStop()
		return nil
	})

	if err := eg.Wait(); err != nil {
		logger.Log.Fatal("Server closed with unexpected error", zap.Error(err))
	}
}
