package grpcutils

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggerUnaryServerInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()

	res, err := handler(ctx, req)
	took := time.Since(start).Round(time.Microsecond)

	if err != nil {
		s := status.Convert(err)

		if s != nil {
			slog.Info(
				"GRPC: Handled unary call",
				"method", info.FullMethod,
				"code", s.Code(),
				"took", took,
			)
		} else {
			slog.Info(
				"GRPC: Handled unary call with error",
				"method", info.FullMethod,
				"took", took,
				"error", err,
			)
		}

	} else {
		slog.Info(
			"GRPC: Handled unary call",
			"method", info.FullMethod,
			"code", codes.OK,
			"took", took,
		)
	}

	return res, err
}

func LoggerStreamServerInterceptor(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	start := time.Now()

	err := handler(srv, ss)
	took := time.Since(start).Round(time.Microsecond)

	if err != nil {
		s := status.Convert(err)

		if s != nil {
			slog.Info(
				"GRPC: Handled stream call",
				"method", info.FullMethod,
				"server_stream", info.IsServerStream,
				"client_stream", info.IsClientStream,
				"code", s.Code(),
				"took", took,
			)
		} else {
			slog.Info(
				"GRPC: Handled stream call with error",
				"method", info.FullMethod,
				"server_stream", info.IsServerStream,
				"client_stream", info.IsClientStream,
				"took", took,
				"error", err,
			)
		}
	} else {
		slog.Info(
			"GRPC: Handled stream call",
			"method", info.FullMethod,
			"server_stream", info.IsServerStream,
			"client_stream", info.IsClientStream,
			"code", codes.OK,
			"took", took,
		)
	}

	return err
}
