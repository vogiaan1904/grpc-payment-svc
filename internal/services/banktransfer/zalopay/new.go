package zalopay

import (
	"net/http"

	bankTf "github.com/vogiaan1904/payment-svc/internal/services/banktransfer"
)

type ZalopayGateway struct {
	OrderTimeoutSeconds         int
	CreateZalopayPaymentLinkURL string
	AppID                       int
	Key1                        string
	Key2                        string
	CallbackErrorCode           int
	HttpClient                  *http.Client
	Host                        string
}

func New(appID int, key1 string, key2 string, host string) bankTf.PaymentGateway {
	return &ZalopayGateway{
		OrderTimeoutSeconds:         300,
		CreateZalopayPaymentLinkURL: "https://sb-openapi.zalopay.vn/v2/create",
		AppID:                       appID,
		Key1:                        key1,
		Key2:                        key2,
		CallbackErrorCode:           -1,
		HttpClient:                  &http.Client{},
		Host:                        host,
	}
}
