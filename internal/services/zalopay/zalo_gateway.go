package zalopay

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vogiaan1904/payment-svc/internal/models"
	service "github.com/vogiaan1904/payment-svc/internal/services"
	"github.com/vogiaan1904/payment-svc/protogen/golang/payment"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Change the type to be exported by capitalizing it
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

// NewZalopayGateway creates a new Zalopay gateway
func NewZalopayGateway(cfg ZalopayConfig) service.PaymentGatewayInterface {
	return &ZalopayGateway{
		OrderTimeoutSeconds:         30,
		CreateZalopayPaymentLinkURL: "https://sb-openapi.zalopay.vn/v2/create",
		AppID:                       cfg.AppID,
		Key1:                        cfg.Key1,
		Key2:                        cfg.Key2,
		CallbackErrorCode:           -1,
		HttpClient:                  &http.Client{},
		Host:                        cfg.Host,
	}
}

func (z *ZalopayGateway) initZaloPayRequestConfig(data ZaloPayRequestConfigInterface) ZaloPayRequestConfig {
	now := time.Now()
	transID := now.Format("060102") // YY MM DD format

	// Append bookingCode to returnUrl if it's not already there
	returnURL := data.ReturnURL
	if strings.Contains(returnURL, "?") {
		returnURL += fmt.Sprintf("&bookingCode=%s", data.OrderCode)
	} else {
		returnURL += fmt.Sprintf("?bookingCode=%s", data.OrderCode)
	}

	// Create embed data
	embedDataObj := embedData{
		RedirectURL: returnURL,
	}
	embedDataJSON, _ := json.Marshal(embedDataObj)

	// Create empty item array
	itemJSON, _ := json.Marshal([]interface{}{})

	// Create config
	config := ZaloPayRequestConfig{
		AppID:              strconv.Itoa(z.AppID),
		AppUser:            "user123",
		AppTime:            now.UnixMilli(),
		Amount:             data.Amount,
		AppTransID:         fmt.Sprintf("%s_%s", transID, data.OrderCode),
		EmbedData:          string(embedDataJSON),
		ExpireDurationSecs: z.OrderTimeoutSeconds,
		Description:        data.Description,
		BankCode:           "",
		CallbackURL:        data.Host + "/zalopay/callback",
		Item:               string(itemJSON),
		Mac:                "",
	}

	// Create MAC
	macInput := fmt.Sprintf("%s|%s|%s|%d|%d|%s|%s",
		config.AppID,
		config.AppTransID,
		config.AppUser,
		config.Amount,
		config.AppTime,
		config.EmbedData,
		config.Item,
	)

	h := hmac.New(sha256.New, []byte(z.Key1))
	h.Write([]byte(macInput))
	config.Mac = hex.EncodeToString(h.Sum(nil))

	return config
}

// ProcessPayment implements the PaymentGatewayInterface
func (g *ZalopayGateway) ProcessPayment(ctx context.Context, req *payment.ProcessPaymentRequest) (*payment.ProcessPaymentResponse, error) {
	data := g.initZaloPayRequestConfig(ZaloPayRequestConfigInterface{
		OrderCode:   req.OrderCode,
		Amount:      int64(req.Amount),
		Description: "E-Commerce",
		ReturnURL:   req.Metadata["return_url"],
		Host:        g.Host,
	})

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.CreateZalopayPaymentLinkURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.HttpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	type zaloPayResponse struct {
		ReturnCode int    `json:"return_code"`
		OrderURL   string `json:"order_url"`
	}

	var zaloResp zaloPayResponse
	if err := json.NewDecoder(resp.Body).Decode(&zaloResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if zaloResp.ReturnCode != 1 {
		return nil, fmt.Errorf("zalopay error: return_code=%d", zaloResp.ReturnCode)
	}

	return &payment.ProcessPaymentResponse{
		PaymentUrl: zaloResp.OrderURL,
		Payment: &payment.PaymentData{
			Id:          data.AppTransID,
			OrderCode:   req.OrderCode,
			Amount:      float64(data.Amount),
			Status:      payment.PaymentStatus_PAYMENT_STATUS_PENDING,
			Method:      payment.PaymentMethod_PAYMENT_METHOD_BANK_TRANSFER,
			GatewayName: string(models.GatewayTypeZalopay),
			Description: data.Description,
			CreatedAt:   time.Now().Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		},
	}, nil
}

func (g *ZalopayGateway) HandleCallback(ctx context.Context, callbackData interface{}) (string, error) {
	zpCallbackData, ok := callbackData.(ZalopayCallbackData)
	if !ok {
		return "", fmt.Errorf("invalid callback data")
	}

	h := hmac.New(sha256.New, []byte(g.Key2))
	h.Write([]byte(zpCallbackData.Data))
	requestMac := hex.EncodeToString(h.Sum(nil))

	if requestMac != zpCallbackData.Mac {
		return "", fmt.Errorf("invalid mac")
	}

	var transData TransactionData
	if err := json.Unmarshal([]byte(zpCallbackData.Data), &transData); err != nil {
		return "", fmt.Errorf("failed to parse transaction data: %w", err)
	}

	parts := strings.Split(transData.AppTransID, "_")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid app_trans_id format")
	}
	fmt.Printf("TransData: %+v\n", transData)

	//TODO: store transaction data to database

	return transData.AppTransID, nil
}

func (g *ZalopayGateway) GetPaymentStatus(ctx context.Context, req *payment.GetPaymentStatusRequest) (*payment.GetPaymentStatusResponse, error) {
	macData := fmt.Sprintf("%d|%s|%s", g.AppID, req.PaymentId, g.Key1)

	h := hmac.New(sha256.New, []byte(g.Key1))
	h.Write([]byte(macData))
	mac := hex.EncodeToString(h.Sum(nil))

	requestBody := map[string]interface{}{
		"app_id":       g.AppID,
		"app_trans_id": req.PaymentId,
		"mac":          mac,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, "POST", "https://sb-openapi.zalopay.vn/v2/query", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := g.HttpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer response.Body.Close()

	var zaloResp zaloPayStatusResponse
	if err := json.NewDecoder(response.Body).Decode(&zaloResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var status payment.PaymentStatus
	switch {
	case zaloResp.ReturnCode == 1 && zaloResp.SubReturnCode == 1:
		status = payment.PaymentStatus_PAYMENT_STATUS_COMPLETED
	case zaloResp.IsProcessing == 1 || zaloResp.ReturnCode == 2:
		status = payment.PaymentStatus_PAYMENT_STATUS_PENDING
	case zaloResp.ReturnCode == 2:
		status = payment.PaymentStatus_PAYMENT_STATUS_FAILED
	default:
		status = payment.PaymentStatus_PAYMENT_STATUS_FAILED
	}

	return &payment.GetPaymentStatusResponse{
		Payment: &payment.PaymentData{
			Id:        req.PaymentId,
			OrderCode: zaloResp.ZpTransID,
			Amount:    float64(zaloResp.Amount),
			Status:    status,
		},
	}, nil
}

func (g *ZalopayGateway) CancelPayment(ctx context.Context, req *payment.CancelPaymentRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
