package grpcutils

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func ErrorUnaryClientInterceptor(f func(s string) error) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) (err error) {
		start := time.Now()

		err = invoker(ctx, method, req, reply, cc)
		if err != nil {
			s := status.Convert(err)
			desc := s.Message()

			newErr := f(desc)
			if newErr != nil {
				err = newErr
			}

			slog.Info(
				"GRPC: Got error in unary call",
				"method", method,
				"code", s.Code(),
				"took", time.Since(start).Round(time.Microsecond),
				"error", desc,
			)
		}

		return
	}
}

func ErrorStreamClientInterceptor(f func(s string) error) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (stream grpc.ClientStream, err error) {
		start := time.Now()

		stream, err = streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			s := status.Convert(err)
			errdesc := s.Message()

			newErr := f(errdesc)
			if newErr != nil {
				err = newErr
			}

			slog.Info(
				"GRPC: Got error in stream call",
				"method", method,
				"server_stream", desc.ServerStreams,
				"client_stream", desc.ClientStreams,
				"code", s.Code(),
				"took", time.Since(start).Round(time.Microsecond),
				"error", errdesc,
			)
		}

		return
	}
}
