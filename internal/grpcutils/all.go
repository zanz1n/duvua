package grpcutils

import "google.golang.org/grpc"

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
