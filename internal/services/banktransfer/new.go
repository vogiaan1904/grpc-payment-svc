package banktransfer

import (
	"sync"

	"github.com/vogiaan1904/payment-svc/internal/models"
)

type GatewayFactory struct {
	gateways map[models.GatewayType]PaymentGateway
	mu       sync.Mutex
}

func NewPaymentGatewayFactory() *GatewayFactory {
	return &GatewayFactory{
		gateways: make(map[models.GatewayType]PaymentGateway),
	}
}
