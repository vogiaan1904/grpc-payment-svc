package service

import (
	"context"

	"github.com/vogiaan1904/payment-svc/protogen/golang/payment"
)

type PaymentGatewayInterface interface {
	ProcessPayment(ctx context.Context, req *payment.ProcessPaymentRequest) (*payment.ProcessPaymentResponse, error)
	GetPaymentStatus(ctx context.Context, req *payment.GetPaymentStatusRequest) (*payment.GetPaymentStatusResponse, error)
	HandleCallback(ctx context.Context, data interface{}) error
}
