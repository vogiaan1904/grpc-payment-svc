package service

import (
	"context"
	"time"

	"github.com/vogiaan1904/payment-svc/internal/models"
	repository "github.com/vogiaan1904/payment-svc/internal/repositories"
	"github.com/vogiaan1904/payment-svc/pkg/log"
	payment "github.com/vogiaan1904/payment-svc/protogen/golang/payment"
	"google.golang.org/protobuf/types/known/emptypb"
)

type implPaymentService struct {
	l    log.Logger
	repo repository.PaymentRepository
	payment.UnimplementedPaymentServiceServer
}

func NewPaymentService(l log.Logger, repo repository.PaymentRepository) payment.PaymentServiceServer {
	return &implPaymentService{
		l:    l,
		repo: repo,
	}
}

func (svc *implPaymentService) ProcessPayment(ctx context.Context, req *payment.ProcessPaymentRequest) (*payment.ProcessPaymentResponse, error) {
	svc.l.Infof(ctx, "Processing payment for order %s", req.OrderId)

	// This is where you'd implement the gateway-specific payment processing
	// For now, we'll just create a pending payment record

	method := models.PaymentMethod(req.Method.String())
	p, err := svc.repo.CreatePayment(ctx, repository.CreatePaymentOptions{
		OrderID:     req.OrderId,
		UserID:      req.UserId,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Method:      method,
		Description: req.Description,
		Metadata:    req.Metadata,
		Status:      models.PaymentStatusPending,
	})
	if err != nil {
		svc.l.Errorf(ctx, "Failed to create payment record: %v", err)
		return nil, err
	}

	// Convert the payment model to protobuf message
	paymentData := &payment.PaymentData{
		Id:               p.ID.Hex(),
		OrderId:          p.OrderID,
		UserId:           p.UserID,
		Amount:           p.Amount,
		Currency:         p.Currency,
		Status:           svc.mapStatusToProto(p.Status),
		Method:           svc.mapMethodToProto(p.Method),
		GatewayReference: p.GatewayReference,
		Description:      p.Description,
		CreatedAt:        p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        p.UpdatedAt.Format(time.RFC3339),
	}

	// In a real implementation, you might return a redirect URL for some payment methods
	return &payment.ProcessPaymentResponse{
		Payment:    paymentData,
		PaymentUrl: "", // Would be populated for redirect-based payment flows
	}, nil
}

func (svc *implPaymentService) GetPaymentStatus(ctx context.Context, req *payment.GetPaymentStatusRequest) (*payment.GetPaymentStatusResponse, error) {
	svc.l.Infof(ctx, "Getting payment status for payment %s", req.PaymentId)

	p, err := svc.repo.FindOnePayment(ctx, repository.FindOnePaymentOptions{
		ID: req.PaymentId,
		FindFilter: repository.FindFilter{
			IncludeDeleted: false,
		},
	})
	if err != nil {
		svc.l.Errorf(ctx, "Failed to get payment: %v", err)
		return nil, err
	}

	paymentData := &payment.PaymentData{
		Id:               p.ID.Hex(),
		OrderId:          p.OrderID,
		UserId:           p.UserID,
		Amount:           p.Amount,
		Currency:         p.Currency,
		Status:           svc.mapStatusToProto(p.Status),
		Method:           svc.mapMethodToProto(p.Method),
		GatewayReference: p.GatewayReference,
		Description:      p.Description,
		CreatedAt:        p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        p.UpdatedAt.Format(time.RFC3339),
	}

	return &payment.GetPaymentStatusResponse{
		Payment: paymentData,
	}, nil
}

func (svc *implPaymentService) CancelPayment(ctx context.Context, req *payment.CancelPaymentRequest) (*emptypb.Empty, error) {
	svc.l.Infof(ctx, "Cancelling payment %s", req.PaymentId)

	// This is where you'd implement gateway-specific cancellation logic
	// For now, we'll just update the status

	err := svc.repo.CancelPayment(ctx, repository.CancelPaymentOptions{
		ID:     req.PaymentId,
		Reason: req.Reason,
	})
	if err != nil {
		svc.l.Errorf(ctx, "Failed to cancel payment: %v", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Helper methods to map between domain models and protobuf types

func (svc *implPaymentService) mapStatusToProto(status models.PaymentStatus) payment.PaymentStatus {
	switch status {
	case models.PaymentStatusPending:
		return payment.PaymentStatus_PAYMENT_STATUS_PENDING
	case models.PaymentStatusCompleted:
		return payment.PaymentStatus_PAYMENT_STATUS_COMPLETED
	case models.PaymentStatusFailed:
		return payment.PaymentStatus_PAYMENT_STATUS_FAILED
	case models.PaymentStatusCancelled:
		return payment.PaymentStatus_PAYMENT_STATUS_CANCELLED
	case models.PaymentStatusRefunded:
		return payment.PaymentStatus_PAYMENT_STATUS_REFUNDED
	default:
		return payment.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}

func (svc *implPaymentService) mapMethodToProto(method models.PaymentMethod) payment.PaymentMethod {
	switch method {
	case models.PaymentMethodCreditCard:
		return payment.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD
	case models.PaymentMethodPaypal:
		return payment.PaymentMethod_PAYMENT_METHOD_PAYPAL
	case models.PaymentMethodBankTransfer:
		return payment.PaymentMethod_PAYMENT_METHOD_BANK_TRANSFER
	case models.PaymentMethodCrypto:
		return payment.PaymentMethod_PAYMENT_METHOD_CRYPTO
	default:
		return payment.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED
	}
}
