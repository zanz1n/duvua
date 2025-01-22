package main

import (
	"log"
	"log/slog"
	"time"

	"github.com/zanz1n/duvua/internal/grpcutils"
	"github.com/zanz1n/duvua/pkg/grpcpool"
	playerpb "github.com/zanz1n/duvua/pkg/pb/player"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

func connectToPlayerGrpc() (*grpcpool.Pool, func()) {
	start := time.Now()

	cfg := GetConfig()

	passwd := cfg.Player.Password
	pool, err := grpcpool.New(
		10,
		cfg.Player.ApiURL,
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.DefaultConfig,
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpcutils.AllUnaryClientInterceptors(playerpb.ConvertError, passwd)...,
		),
		grpc.WithChainStreamInterceptor(
			grpcutils.AllStreamClientInterceptors(playerpb.ConvertError, passwd)...,
		),
	)

	if err != nil {
		log.Fatalln("Failed to connect to player grpc server:", err)
	}

	slog.Info(
		"Connected to player GRPC server",
		"took", time.Since(start).Round(time.Millisecond),
	)

	return pool, func() {
		start := time.Now()
		pool.Close()
		slog.Info(
			"Closed GRPC player connection pool",
			"took", time.Since(start).Round(time.Millisecond),
		)
	}
}

func connectToDavinciGrpc() (*grpcpool.Pool, func()) {
	start := time.Now()

	cfg := GetConfig()

	passwd := cfg.Welcomer.Password
	pool, err := grpcpool.New(
		10,
		cfg.Welcomer.ApiURL,
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.DefaultConfig,
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpcutils.AllUnaryClientInterceptors(nilErrConverter, passwd)...,
		),
		grpc.WithChainStreamInterceptor(
			grpcutils.AllStreamClientInterceptors(nilErrConverter, passwd)...,
		),
	)

	if err != nil {
		log.Fatalln("Failed to connect to davinci grpc server:", err)
	}

	slog.Info(
		"Connected to davinci GRPC server",
		"took", time.Since(start).Round(time.Millisecond),
	)

	return pool, func() {
		start := time.Now()
		pool.Close()
		slog.Info(
			"Closed GRPC davinci connection pool",
			"took", time.Since(start).Round(time.Millisecond),
		)
	}
}

func nilErrConverter(msg string) error {
	return nil
}
