package grpcutils

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RecoverUnaryServerInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (res any, err error) {
	start := time.Now()

	defer func() {
		if err2 := recover(); err2 != nil {
			fmterr := fmt.Sprint(err2)
			slog.Error(
				"GRPC: PANIC in unary call",
				"method", info.FullMethod,
				"took", time.Since(start).Round(time.Microsecond),
				"error", fmterr,
			)

			err = status.Error(codes.Internal, fmterr)
		}
	}()

	res, err = handler(ctx, req)
	return
}

func RecoverStreamServerInterceptor(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) (err error) {
	start := time.Now()

	defer func() {
		if err2 := recover(); err2 != nil {
			fmterr := fmt.Sprint(err2)
			slog.Error(
				"GRPC: PANIC in stream call",
				"method", info.FullMethod,
				"took", time.Since(start).Round(time.Microsecond),
				"error", fmterr,
			)

			err = status.Error(codes.Internal, fmterr)
		}
	}()

	err = handler(srv, ss)
	return
}
