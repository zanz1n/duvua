package grpcutils

import "google.golang.org/grpc"

func AllUnaryClientInterceptors() []grpc.UnaryClientInterceptor {
	return []grpc.UnaryClientInterceptor{
		LoggerUnaryClientInterceptor,
		ErrorUnaryClientInterceptor,
	}
}

func AllStreamClientInterceptors() []grpc.StreamClientInterceptor {
	return []grpc.StreamClientInterceptor{
		LoggerStreamClientInterceptor,
		ErrorStreamClientInterceptor,
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
