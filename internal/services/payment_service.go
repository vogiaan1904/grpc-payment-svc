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

	return &payment.ProcessPaymentResponse{
		PaymentUrl: pRes.PaymentUrl,
		Payment:    pRes.Payment,
	}, nil
}

func (svc *implPaymentService) GetPaymentStatus(ctx context.Context, req *payment.GetPaymentStatusRequest) (*payment.GetPaymentStatusResponse, error) {
	gw, err := svc.gatewayFactory.GetGateway(models.GatewayType(req.GatewayName))
	if err != nil {
		svc.l.Errorf(ctx, "failed to get payment gateway: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid gateway: %v", err)
	}

}

func (svc *implPaymentService) CancelPayment(ctx context.Context, req *payment.CancelPaymentRequest) (*emptypb.Empty, error) {
	gw, err := svc.gatewayFactory.GetGateway(models.GatewayType(req.GatewayName))
	return &emptypb.Empty{}, nil
}
