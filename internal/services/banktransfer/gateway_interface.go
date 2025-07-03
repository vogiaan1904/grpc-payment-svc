package banktransfer

import (
	"context"

	"github.com/vogiaan1904/payment-svc/protogen/golang/payment"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PaymentGateway interface {
	ProcessPayment(ctx context.Context, req *payment.ProcessPaymentRequest) (*payment.ProcessPaymentResponse, error)
	HandleCallback(ctx context.Context, data interface{}) (string, error)
	CancelPayment(ctx context.Context, req *payment.CancelPaymentRequest) (*emptypb.Empty, error)
}
