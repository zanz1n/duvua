package grpcutils

import (
	"context"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var validate = validator.New()

func ValidateServerUnaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	if err := validate.Struct(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return handler(ctx, req)
}
