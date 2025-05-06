package grpcservices

import (
	"context"
	"log"

	"github.com/vogiaan1904/payment-svc/internal/interceptors"
	pkgLog "github.com/vogiaan1904/payment-svc/pkg/log"
	"github.com/vogiaan1904/payment-svc/protogen/golang/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClients struct {
	Order order.OrderServiceClient
}

type cleanupFunc func()

func InitGrpcClients(orderAddr string, logger pkgLog.Logger, redactedFields []string) (*GrpcClients, cleanupFunc, error) {
	ctx := context.Background()

	var cleanupFuncs []cleanupFunc
	clients := &GrpcClients{}

	loggingInterceptor := interceptors.GrpcClientLoggingInterceptor(logger, redactedFields)

	orderConn, err := grpc.NewClient(
		orderAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(loggingInterceptor),
	)
	if err != nil {
		return nil, nil, err
	}
	clients.Order = order.NewOrderServiceClient(orderConn)

	// Add more clients here...
	cleanupFuncs = append(cleanupFuncs, func() {
		if err := orderConn.Close(); err != nil {
			log.Printf("failed to close order gRPC connection: %v", err)
		}
	})

	cleanupFunc := func() {
		for _, fn := range cleanupFuncs {
			fn()
		}
		logger.Info(ctx, "gRPC clients cleaned up")
	}

	logger.Info(ctx, "gRPC clients initialized")

	return clients, cleanupFunc, nil
}
