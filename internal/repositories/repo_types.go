package repository

import (
	"github.com/vogiaan1904/payment-svc/internal/models"
)

// Common options
type FindFilter struct {
	IncludeDeleted bool
}

// Payment Repository Options
type CreatePaymentOptions struct {
	OrderID          string
	UserID           string
	Amount           float64
	Currency         string
	Method           models.PaymentMethod
	Description      string
	Metadata         map[string]string
	GatewayReference string
	Status           models.PaymentStatus
}

type FindOnePaymentOptions struct {
	ID         string
	OrderID    string
	UserID     string
	FindFilter FindFilter
}

type UpdatePaymentStatusOptions struct {
	ID               string
	Status           models.PaymentStatus
	GatewayReference string
}

type CancelPaymentOptions struct {
	ID     string
	Reason string
}
