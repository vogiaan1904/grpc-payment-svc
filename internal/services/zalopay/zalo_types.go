package zalopay

// ZalopayGateway struct is now defined in zalo_gateway.go

type ZalopayConfig struct {
	AppID int
	Key1  string
	Key2  string
	Host  string
}

type ZaloPayRequestConfigInterface struct {
	OrderCode   string
	Amount      int64
	Description string
	ReturnURL   string
	Host        string
}

type ZaloPayCallbackResponse struct {
	ReturnCode    int    `json:"return_code"`
	ReturnMessage string `json:"return_message"`
}

type TransactionData struct {
	AppTransID string `json:"app_trans_id"`
}

type ZalopayCallbackData struct {
	Data string `json:"data"`
	Mac  string `json:"mac"`
}

type CallbackResult struct {
	Success   bool
	Response  ZaloPayCallbackResponse
	OrderCode string
}

type embedData struct {
	RedirectURL string `json:"redirecturl"`
}

type ZaloPayRequestConfig struct {
	AppID              string `json:"app_id"`
	AppUser            string `json:"app_user"`
	AppTime            int64  `json:"app_time"`
	Amount             int64  `json:"amount"`
	AppTransID         string `json:"app_trans_id"`
	EmbedData          string `json:"embed_data"`
	ExpireDurationSecs int    `json:"expire_duration_seconds"`
	Description        string `json:"description"`
	BankCode           string `json:"bank_code"`
	CallbackURL        string `json:"callback_url"`
	Item               string `json:"item"`
	Mac                string `json:"mac"`
}

type zaloPayStatusResponse struct {
	ReturnCode       int    `json:"return_code"`
	ReturnMessage    string `json:"return_message"`
	SubReturnCode    int    `json:"sub_return_code"`
	SubReturnMessage string `json:"sub_return_message"`
	IsProcessing     int    `json:"is_processing"`
	Amount           int64  `json:"amount"`
	ZpTransID        string `json:"zp_trans_id"`
}
