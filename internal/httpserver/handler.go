package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/vogiaan1904/payment-svc/internal/models"
	service "github.com/vogiaan1904/payment-svc/internal/services"
	zpGW "github.com/vogiaan1904/payment-svc/internal/services/zalopay"
)

func (s *Server) handleZalopayCallback(w http.ResponseWriter, r *http.Request) {
	var callbackData zpGW.ZalopayCallbackData
	if err := json.NewDecoder(r.Body).Decode(&callbackData); err != nil {
		s.logger.Errorf(r.Context(), "Failed to decode callback data: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := service.HandlePaymentCallback(s.paymentSvc, r.Context(), callbackData, models.GatewayTypeZalopay); err != nil {
		s.logger.Errorf(r.Context(), "Failed to process callback: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
