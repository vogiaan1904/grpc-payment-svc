package banktransfer

import (
	"fmt"

	"github.com/vogiaan1904/payment-svc/internal/models"
)

func (f *GatewayFactory) RegisterGateway(gatewayType models.GatewayType, gateway PaymentGateway) error {
	if gateway == nil {
		return fmt.Errorf("gateway cannot be nil")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.gateways[gatewayType]; exists {
		return fmt.Errorf("gateway %s already registered", gatewayType)
	}

	f.gateways[gatewayType] = gateway
	return nil
}

func (f *GatewayFactory) GetGateway(gatewayType models.GatewayType) (PaymentGateway, error) {
	gateway, exists := f.gateways[gatewayType]
	if !exists {
		return nil, fmt.Errorf("unsupported gateway type: %s", gatewayType)
	}
	return gateway, nil
}
