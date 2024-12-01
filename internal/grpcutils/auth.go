package grpcutils

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	errMissingAuthorization = status.Errorf(
		codes.Unauthenticated,
		"no incoming `authorization` metadata in grpc context",
	)
	errPasswordMismatches = status.Errorf(
		codes.Unauthenticated,
		"the `authorization` metadata password mismatches",
	)
)

func AuthUnaryServerInterceptor(passwd string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if passwd != "" && !strings.HasPrefix(info.FullMethod, "/grpc.reflection") {
			auth := metadata.ValueFromIncomingContext(ctx, "authorization")
			if len(auth) != 1 {
				return nil, errMissingAuthorization
			}

			if auth[0] != passwd {
				return nil, errPasswordMismatches
			}

		}

		return handler(ctx, req)
	}
}

func AuthStreamServerInterceptor(passwd string) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if passwd != "" && !strings.HasPrefix(info.FullMethod, "/grpc.reflection") {
			auth := metadata.ValueFromIncomingContext(ss.Context(), "authorization")
			if len(auth) != 1 {
				return errMissingAuthorization
			}

			if auth[0] != passwd {
				return errPasswordMismatches
			}

		}

		return handler(srv, ss)
	}
}

func AuthUnaryClientInterceptor(passwd string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if passwd != "" && !strings.HasPrefix(method, "/grpc.reflection") {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", passwd)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func AuthStreamClientInterceptor(passwd string) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		if passwd != "" && !strings.HasPrefix(method, "/grpc.reflection") {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", passwd)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}
