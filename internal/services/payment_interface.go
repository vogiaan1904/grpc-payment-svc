package service

import (
	"context"

	"github.com/vogiaan1904/payment-svc/internal/models"
)

type PaymentServiceInterface interface {
	HandleCallback(ctx context.Context, data interface{}, gatewayType models.GatewayType) error
}
