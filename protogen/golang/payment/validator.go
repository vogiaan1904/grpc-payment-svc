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

func (r *ProcessBankTransferPaymentRequest) Validate() error {
	log.Printf("Validate request: %+v", r)
	if r.OrderCode == "" {
		log.Printf("Order code is required")
		return ErrRequiredField
	}
	if r.Amount <= 0 {
		log.Printf("Amount is required")
		return ErrInvalidInput
	}
	if r.UserId == "" {
		log.Printf("User ID is required")
		return ErrRequiredField
	}
	if r.Provider != string(models.GatewayTypeZalopay) {
		log.Printf("Provider is invalid")
		return ErrInvalidInput
	}

	return nil
}
