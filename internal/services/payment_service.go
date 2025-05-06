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
	var o *order.OrderData
	res, err := svc.orderSvc.FindOne(ctx, &order.FindOneRequest{
		Id: req.OrderId,
	})
	if err != nil {
		svc.l.Errorf(ctx, "payment.orderSvc.FindOne: %v", err)
		return nil, err
	}

	if res != nil {
		o = res.Order
	}

	if o.Status != order.OrderStatus_PENDING {
		svc.l.Errorf(ctx, "payment.ErrOrderNotPending: %v", ErrOrderNotPending)
		return nil, status.Error(codes.FailedPrecondition, ErrOrderNotPending.Error())
	}

	gateway, err := svc.gatewayFactory.GetGateway(models.GatewayType(req.GatewayName))
	if err != nil {
		svc.l.Errorf(ctx, "payment.gatewayFactory.GetGateway: %v", err)
		return nil, err
	}

	// Get client IP from metadata if available
	clientIP := ""
	if req.Metadata != nil {
		if ip, ok := req.Metadata["client_ip"]; ok {
			clientIP = ip
		}
	}

	// Get host and return URL from metadata
	host := ""
	returnURL := ""
	if req.Metadata != nil {
		if h, ok := req.Metadata["host"]; ok {
			host = h
		}
		if url, ok := req.Metadata["return_url"]; ok {
			returnURL = url
		}
	}

	// Process the payment using the selected gateway
	paymentURL, err := gateway.ProcessPayment(ctx, ProcessPaymentOptions{
		Ip:          clientIP,
		Amount:      req.Amount,
		OrderNumber: req.OrderId,
		Host:        host,
		ReturnURL:   returnURL,
	})
	if err != nil {
		svc.l.Errorf(ctx, "payment.gateway.ProcessPayment: %v", err)
		return nil, err
	}

	return &payment.ProcessPaymentResponse{
		PaymentUrl: paymentURL,
	}, nil
}

func (svc *implPaymentService) GetPaymentStatus(ctx context.Context, req *payment.GetPaymentStatusRequest) (*payment.GetPaymentStatusResponse, error) {
	// For demonstration, create a dummy payment data
	paymentData := &payment.PaymentData{
		Id:        req.PaymentId,
		Status:    payment.PaymentStatus_PAYMENT_STATUS_COMPLETED,
		Method:    payment.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD,
		Amount:    1000,
		OrderId:   "order-123",
		UserId:    "user-123",
		CreatedAt: "2023-06-01T12:00:00Z",
		UpdatedAt: "2023-06-01T12:05:00Z",
	}

	return &payment.GetPaymentStatusResponse{
		Payment: paymentData,
	}, nil
}

func (svc *implPaymentService) CancelPayment(ctx context.Context, req *payment.CancelPaymentRequest) (*emptypb.Empty, error) {
	// For simplicity, just return empty response
	return &emptypb.Empty{}, nil
}
