package grpcutils

import (
	"context"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var validate = validator.New()

func init() {
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		tag := field.Tag.Get("protobuf")
		_, next, ok := strings.Cut(tag, "name=")
		if !ok {
			return field.Name
		}

		name, _, ok := strings.Cut(next, ",")
		if !ok {
			return field.Name
		}

		return name
	})
}

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
