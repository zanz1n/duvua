package grpcutils

import "google.golang.org/grpc"

type errh = func(s string) error

func AllUnaryClientInterceptors(errf errh, passwd string) []grpc.UnaryClientInterceptor {
	return []grpc.UnaryClientInterceptor{
		LoggerUnaryClientInterceptor,
		ErrorUnaryClientInterceptor(errf),
		AuthUnaryClientInterceptor(passwd),
	}
}

func AllStreamClientInterceptors(errf errh, passwd string) []grpc.StreamClientInterceptor {
	return []grpc.StreamClientInterceptor{
		LoggerStreamClientInterceptor,
		ErrorStreamClientInterceptor(errf),
		AuthStreamClientInterceptor(passwd),
	}
}

func AllUnaryServerInterceptors(passwd string) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		LoggerUnaryServerInterceptor,
		RecoverUnaryServerInterceptor,
		AuthUnaryServerInterceptor(passwd),
		ValidateServerUnaryInterceptor,
	}
}

func AllStreamServerInterceptors(passwd string) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		LoggerStreamServerInterceptor,
		RecoverStreamServerInterceptor,
		AuthStreamServerInterceptor(passwd),
	}
}
