package payment

import (
	"errors"
	"log"

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
	log.Printf("Validate request: %+v", r)
	if r.OrderCode == "" {
		log.Printf("Order code is required")
		return ErrRequiredField
	}
	if r.Amount <= 0 {
		log.Printf("Invalid amount")
		return ErrInvalidInput
	}
	if r.UserId == "" {
		log.Printf("User ID is required")
		return ErrRequiredField
	}
	if r.Provider != string(models.GatewayTypeZalopay) {
		log.Printf("Invalid provider")
		return ErrInvalidInput
	}

	return nil
}
