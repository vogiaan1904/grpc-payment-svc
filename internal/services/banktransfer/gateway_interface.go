package banktransfer

import (
	"context"

	"github.com/vogiaan1904/payment-svc/protogen/golang/payment"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PaymentGatewayInterface interface {
	ProcessPayment(ctx context.Context, req *payment.ProcessBankTransferPaymentRequest) (*payment.ProcessBankTransferPaymentResponse, error)
	HandleCallback(ctx context.Context, data interface{}) (string, error)
	CancelPayment(ctx context.Context, req *payment.CancelBankTransferPaymentRequest) (*emptypb.Empty, error)
}
