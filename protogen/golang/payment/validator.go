package payment

import (
	"errors"

	"github.com/vogiaan1904/payment-svc/internal/models"
)

var WarnErrors = []error{
	ErrInvalidInput,
	ErrRequiredField,
}

var (
	ErrRequiredField = errors.New("required field is missing")
	ErrInvalidInput  = errors.New("invalid input")
)

func (r *ProcessPaymentRequest) Validate() error {
	if r.OrderId == "" {
		return ErrRequiredField
	}
	if r.Amount <= 0 {
		return ErrInvalidInput
	}
	if r.Method == 0 {
		return ErrRequiredField
	}
	if r.UserId == "" {
		return ErrRequiredField
	}
	if r.GatewayName != "" && r.GatewayName != string(models.GatewayTypeZalopay) {
		return ErrInvalidInput
	}

	return nil
}
