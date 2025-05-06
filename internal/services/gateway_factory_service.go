package service

import (
	"fmt"

	"github.com/vogiaan1904/payment-svc/internal/models"
)

type PaymentGatewayFactory struct {
	gateways map[models.GatewayType]PaymentGatewayInterface
}

func NewPaymentGatewayFactory() *PaymentGatewayFactory {
	return &PaymentGatewayFactory{
		gateways: make(map[models.GatewayType]PaymentGatewayInterface),
	}
}

func (f *PaymentGatewayFactory) RegisterGateway(gatewayType models.GatewayType, gateway PaymentGatewayInterface) {
	f.gateways[gatewayType] = gateway
}

func (f *PaymentGatewayFactory) GetGateway(gatewayType models.GatewayType) (PaymentGatewayInterface, error) {
	gateway, exists := f.gateways[gatewayType]
	if !exists {
		return nil, fmt.Errorf("unsupported gateway type: %s", gatewayType)
	}
	return gateway, nil
}
