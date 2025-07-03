package banktransfer

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

type implPaymentService struct {
	l        log.Logger
	gwf      *GatewayFactory
	orderSvc order.OrderServiceClient
	temporal client.Client
	payment.UnimplementedPaymentServiceServer
}

func NewPaymentService(l log.Logger, gwf *GatewayFactory, orderSvc order.OrderServiceClient, temporal client.Client) payment.PaymentServiceServer {
	return &implPaymentService{
		l:        l,
		gwf:      gwf,
		orderSvc: orderSvc,
		temporal: temporal,
	}
}

func (svc *implPaymentService) ProcessPayment(ctx context.Context, req *payment.ProcessPaymentRequest) (*payment.ProcessPaymentResponse, error) {
	res, err := svc.orderSvc.FindOne(ctx, &order.FindOneRequest{Request: &order.FindOneRequest_Code{Code: req.OrderCode}})
	if err != nil {
		svc.l.Warnf(ctx, "failed to find order: %v", err)
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	if res == nil || res.Order == nil {
		svc.l.Warnf(ctx, "order not found: %v", ErrOrderNotFound)
		return nil, status.Error(codes.NotFound, ErrOrderNotFound.Error())
	}

	if res.Order.Status != order.OrderStatus_PAYMENT_PENDING {
		svc.l.Warnf(ctx, "order status validation failed: %v", ErrOrderNotPending)
		return nil, status.Error(codes.FailedPrecondition, ErrOrderNotPending.Error())
	}

	gw, err := svc.gwf.GetGateway(models.GatewayType(req.Provider))
	if err != nil {
		svc.l.Warnf(ctx, "failed to get payment gateway: %v", err)
		return nil, status.Error(codes.InvalidArgument, ErrInvalidGateway.Error())
	}

	pRes, err := gw.ProcessPayment(ctx, req)
	if err != nil {
		svc.l.Errorf(ctx, "payment processing failed: %v", err)
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return pRes, nil
}

func (svc *implPaymentService) CancelPayment(ctx context.Context, req *payment.CancelPaymentRequest) (*emptypb.Empty, error) {
	resp, err := svc.orderSvc.FindOne(ctx, &order.FindOneRequest{Request: &order.FindOneRequest_Code{Code: req.GetOrderCode()}})
	if err != nil {
		svc.l.Errorf(ctx, "failed to find order: %v", err)
		return nil, status.Errorf(codes.Internal, "error retrieving order: %v", err)
	}

	gw, err := svc.gwf.GetGateway(models.GatewayType(resp.Order.Provider))
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
	gw, err := svc.gwf.GetGateway(gatewayType)
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
		TaskQueue:                TaskQueueName,
		WorkflowExecutionTimeout: time.Hour * 24,
		WorkflowRunTimeout:       time.Hour * 24,
		WorkflowTaskTimeout:      time.Minute * 1,
	}

	svc.l.Infof(ctx, "Starting workflow with ID: %s", wfID)
	we, err := svc.temporal.ExecuteWorkflow(ctx, wfOpts, WorkflowName, wfParams)
	if err != nil {
		svc.l.Errorf(ctx, "Failed to start workflow: %v", err)
		return status.Errorf(codes.Internal, "failed to initiate order processing: %v", err)
	}

	svc.l.Infof(ctx, "Workflow started successfully. WorkflowID: %s, RunID: %s", we.GetID(), we.GetRunID())
	return nil
}

func HandlePaymentCallback(svc payment.PaymentServiceServer, ctx context.Context, data interface{}, gatewayType models.GatewayType) error {
	impl, ok := svc.(*implPaymentService)
	if !ok {
		return status.Errorf(codes.Internal, "invalid payment service implementation")
	}
	return impl.HandleCallback(ctx, data, gatewayType)
}
