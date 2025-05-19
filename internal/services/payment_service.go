package service

import (
	"context"

	"github.com/vogiaan1904/payment-svc/internal/models"
	"github.com/vogiaan1904/payment-svc/pkg/log"
	"github.com/vogiaan1904/payment-svc/protogen/golang/order"
	"github.com/vogiaan1904/payment-svc/protogen/golang/payment"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type implPaymentService struct {
	l              log.Logger
	gatewayFactory *PaymentGatewayFactory
	orderSvc       order.OrderServiceClient
	payment.UnimplementedPaymentServiceServer
}

func NewPaymentService(l log.Logger, factory *PaymentGatewayFactory, orderSvc order.OrderServiceClient) payment.PaymentServiceServer {
	return &implPaymentService{
		l:              l,
		gatewayFactory: factory,
		orderSvc:       orderSvc,
	}
}

func (svc *implPaymentService) ProcessPayment(ctx context.Context, req *payment.ProcessPaymentRequest) (*payment.ProcessPaymentResponse, error) {
	res, err := svc.orderSvc.FindOne(ctx, &order.FindOneRequest{Id: req.OrderId})
	if err != nil {
		svc.l.Errorf(ctx, "failed to find order: %v", err)
		return nil, status.Errorf(codes.Internal, "error retrieving order: %v", err)
	}

	if res == nil || res.Order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	if res.Order.Status != order.OrderStatus_PENDING {
		svc.l.Errorf(ctx, "order status validation failed: %v", ErrOrderNotPending)
		return nil, status.Error(codes.FailedPrecondition, ErrOrderNotPending.Error())
	}

	gw, err := svc.gatewayFactory.GetGateway(models.GatewayType(req.GatewayName))
	if err != nil {
		svc.l.Errorf(ctx, "failed to get payment gateway: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid gateway: %v", err)
	}

	pRes, err := gw.ProcessPayment(ctx, req)
	if err != nil {
		svc.l.Errorf(ctx, "payment processing failed: %v", err)
		return nil, status.Errorf(codes.Internal, "payment processing failed: %v", err)
	}

	return pRes, nil
}

func (svc *implPaymentService) GetPaymentStatus(ctx context.Context, req *payment.GetPaymentStatusRequest) (*payment.GetPaymentStatusResponse, error) {
	gw, err := svc.gatewayFactory.GetGateway(models.GatewayType(req.GatewayName))
	if err != nil {
		svc.l.Errorf(ctx, "failed to get payment gateway: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid gateway: %v", err)
	}

	pRes, err := gw.GetPaymentStatus(ctx, req)
	if err != nil {
		svc.l.Errorf(ctx, "failed to get payment status: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get payment status: %v", err)
	}

	return pRes, nil
}

func (svc *implPaymentService) CancelPayment(ctx context.Context, req *payment.CancelPaymentRequest) (*emptypb.Empty, error) {
	gw, err := svc.gatewayFactory.GetGateway(models.GatewayType(req.GatewayName))
	if err != nil {
		svc.l.Errorf(ctx, "failed to get payment gateway: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid gateway: %v", err)
	}

	cRes, err := gw.CancelPayment(ctx, req)
	if err != nil {
		svc.l.Errorf(ctx, "failed to cancel payment: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to cancel payment: %v", err)
	}

	return cRes, nil
}

// HandleCallback processes payment gateway callbacks
// This is for HTTP callbacks, not part of the gRPC service definition
func (svc *implPaymentService) HandleCallback(ctx context.Context, data interface{}, gatewayType models.GatewayType) error {
	gw, err := svc.gatewayFactory.GetGateway(gatewayType)
	if err != nil {
		svc.l.Errorf(ctx, "failed to get payment gateway: %v", err)
		return status.Errorf(codes.InvalidArgument, "invalid gateway: %v", err)
	}

	return gw.HandleCallback(ctx, data)
}

// Export the HandleCallback function to be used with HTTP callbacks
func HandlePaymentCallback(svc payment.PaymentServiceServer, ctx context.Context, data interface{}, gatewayType models.GatewayType) error {
	// Type assertion to get the concrete implementation
	impl, ok := svc.(*implPaymentService)
	if !ok {
		return status.Errorf(codes.Internal, "invalid payment service implementation")
	}
	return impl.HandleCallback(ctx, data, gatewayType)
}
