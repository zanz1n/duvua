package grpcutils

import "google.golang.org/grpc"

func AllUnaryClientInterceptors(errf func(s string) error) []grpc.UnaryClientInterceptor {
	return []grpc.UnaryClientInterceptor{
		LoggerUnaryClientInterceptor,
		ErrorUnaryClientInterceptor(errf),
	}
}

func AllStreamClientInterceptors(errf func(s string) error) []grpc.StreamClientInterceptor {
	return []grpc.StreamClientInterceptor{
		LoggerStreamClientInterceptor,
		ErrorStreamClientInterceptor(errf),
	}
}

func AllUnaryServerInterceptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		LoggerUnaryServerInterceptor,
		RecoverUnaryServerInterceptor,
		ValidateServerUnaryInterceptor,
	}
}

func AllStreamServerInterceptors() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		LoggerStreamServerInterceptor,
		RecoverStreamServerInterceptor,
	}
}
