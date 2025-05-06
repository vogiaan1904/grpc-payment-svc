package service

import "context"

type ProcessPaymentOptions struct {
	Ip          string
	Amount      float64
	OrderNumber string
	Host        string
	ReturnURL   string
}

type CallbackData struct {
	Data any
	Host string
}

type PaymentGatewayInterface interface {
	ProcessPayment(ctx context.Context, opt ProcessPaymentOptions) (string, error)
	Callback(ctx context.Context, data CallbackData) error
}
