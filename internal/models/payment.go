package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusCancelled PaymentStatus = "CANCELLED"
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"
)

type PaymentMethod string

const (
	PaymentMethodCreditCard   PaymentMethod = "CREDIT_CARD"
	PaymentMethodPaypal       PaymentMethod = "PAYPAL"
	PaymentMethodBankTransfer PaymentMethod = "BANK_TRANSFER"
	PaymentMethodCrypto       PaymentMethod = "CRYPTO"
)

type GatewayType string

const (
	GatewayTypeZalopay GatewayType = "ZALOPAY"
)

type Payment struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	OrderID          string             `bson:"order_id"`
	UserID           string             `bson:"user_id"`
	Amount           float64            `bson:"amount"`
	Currency         string             `bson:"currency"`
	Status           PaymentStatus      `bson:"status"`
	Method           PaymentMethod      `bson:"method"`
	GatewayReference string             `bson:"gateway_reference"`
	Description      string             `bson:"description"`
	Metadata         map[string]string  `bson:"metadata,omitempty"`
	CreatedAt        time.Time          `bson:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at"`
	DeletedAt        *time.Time         `bson:"deleted_at,omitempty"`
}
