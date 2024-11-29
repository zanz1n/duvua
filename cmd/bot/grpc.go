package main

import (
	"log"
	"log/slog"
	"time"

	"github.com/zanz1n/duvua/config"
	"github.com/zanz1n/duvua/internal/grpcutils"
	"github.com/zanz1n/duvua/pkg/grpcpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

func connectToPlayerGrpc() *grpcpool.Pool {
	start := time.Now()

	cfg := config.GetConfig()

	pool, err := grpcpool.New(
		10,
		cfg.Player.ApiURL,
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.DefaultConfig,
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(grpcutils.AllUnaryClientInterceptors()...),
		grpc.WithChainStreamInterceptor(grpcutils.AllStreamClientInterceptors()...),
	)

	if err != nil {
		log.Fatalln("Failed to connect to player grpc server:", err)
	}

	slog.Info(
		"Connected to player GRPC server",
		"took", time.Since(start).Round(time.Millisecond),
	)

	return pool
}
