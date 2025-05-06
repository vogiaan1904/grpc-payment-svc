package interceptors

import (
	"context"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrorHandlerInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = status.Errorf(codes.Internal, "internal server error: %v\n%s", r, debug.Stack())
		}
	}()
	resp, err = handler(ctx, req)
	if err != nil {
		if _, ok := status.FromError(err); ok {
			return resp, err
		}

		return resp, status.Error(codes.Internal, "internal server error")
	}
	return resp, nil
}
