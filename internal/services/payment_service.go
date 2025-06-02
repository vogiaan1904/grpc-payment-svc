package service

import (
	"context"
	"time"

	"github.com/vogiaan1904/payment-svc/internal/models"
	"github.com/vogiaan1904/payment-svc/pkg/log"
	"github.com/vogiaan1904/payment-svc/protogen/golang/order"
	"github.com/vogiaan1904/payment-svc/protogen/golang/payment"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const PostPaymentOrderTaskQueue = "POST_PAYMENT_ORDER_TASK_QUEUE"

type implPaymentService struct {
	l              log.Logger
	gatewayFactory *PaymentGatewayFactory
	orderSvc       order.OrderServiceClient
	temporalClient client.Client
	payment.UnimplementedPaymentServiceServer
}

func NewPaymentService(l log.Logger, factory *PaymentGatewayFactory, orderSvc order.OrderServiceClient, tCli client.Client) payment.PaymentServiceServer {
	return &implPaymentService{
		l:              l,
		gatewayFactory: factory,
		orderSvc:       orderSvc,
		temporalClient: tCli,
	}
}

func (svc *implPaymentService) ProcessPayment(ctx context.Context, req *payment.ProcessPaymentRequest) (*payment.ProcessPaymentResponse, error) {
	res, err := svc.orderSvc.FindOne(ctx, &order.FindOneRequest{Code: req.OrderCode})
	if err != nil {
		svc.l.Errorf(ctx, "failed to find order: %v", err)
		return nil, status.Errorf(codes.Internal, "error retrieving order: %v", err)
	}

	if res == nil || res.Order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	if res.Order.Status != order.OrderStatus_PAYMENT_PENDING {
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

func (svc *implPaymentService) HandleCallback(ctx context.Context, data interface{}, gatewayType models.GatewayType) error {
	gw, err := svc.gatewayFactory.GetGateway(gatewayType)
	if err != nil {
		svc.l.Errorf(ctx, "failed to get payment gateway: %v", err)
		return status.Errorf(codes.InvalidArgument, "invalid gateway: %v", err)
	}

	oCode, err := gw.HandleCallback(ctx, data)
	if err != nil {
		svc.l.Errorf(ctx, "failed to handle callback: %v", err)
		return status.Errorf(codes.Internal, "failed to handle callback: %v", err)
	}

	wfID := "order_post_payment_" + oCode
	wfParams := OrderWorkflowParams{
		OrderCode: oCode,
	}

	wfOpts := client.StartWorkflowOptions{
		ID:                       wfID,
		TaskQueue:                PostPaymentOrderTaskQueue,
		WorkflowExecutionTimeout: time.Hour * 24,
		WorkflowRunTimeout:       time.Hour * 24,
		WorkflowTaskTimeout:      time.Minute * 1,
	}

	svc.l.Infof(ctx, "Starting OrderProcessingWorkflow with ID: %s", wfID)
	we, err := svc.temporalClient.ExecuteWorkflow(ctx, wfOpts, "ProcessPostPaymentOrder", &wfParams)
	if err != nil {
		svc.l.Errorf(ctx, "Failed to start OrderProcessingWorkflow: %v", err)
		return status.Errorf(codes.Internal, "failed to initiate order processing: %v", err)
	}

	svc.l.Infof(ctx, "OrderProcessingWorkflow started successfully. WorkflowID: %s, RunID: %s", we.GetID(), we.GetRunID())
	return nil
}

func HandlePaymentCallback(svc payment.PaymentServiceServer, ctx context.Context, data interface{}, gatewayType models.GatewayType) error {
	impl, ok := svc.(*implPaymentService)
	if !ok {
		return status.Errorf(codes.Internal, "invalid payment service implementation")
	}
	return impl.HandleCallback(ctx, data, gatewayType)
}
