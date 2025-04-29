package repository

import (
	"context"

	"github.com/vogiaan1904/payment-svc/internal/models"
)

type PaymentRepository interface {
	CreatePayment(ctx context.Context, opt CreatePaymentOptions) (models.Payment, error)
	FindOnePayment(ctx context.Context, opt FindOnePaymentOptions) (models.Payment, error)
	UpdatePaymentStatus(ctx context.Context, opt UpdatePaymentStatusOptions) error
	CancelPayment(ctx context.Context, opt CancelPaymentOptions) error
}
