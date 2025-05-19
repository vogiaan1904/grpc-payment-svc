package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/vogiaan1904/payment-svc/config"
	"github.com/vogiaan1904/payment-svc/internal/interceptors"
	"github.com/vogiaan1904/payment-svc/internal/models"
	service "github.com/vogiaan1904/payment-svc/internal/services"
	zpGW "github.com/vogiaan1904/payment-svc/internal/services/zalopay"
	pkgGrpc "github.com/vogiaan1904/payment-svc/pkg/grpc"
	pkgLog "github.com/vogiaan1904/payment-svc/pkg/log"
	"github.com/vogiaan1904/payment-svc/protogen/golang/payment"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	l := pkgLog.InitializeZapLogger(pkgLog.ZapConfig{
		Level:    cfg.Log.Level,
		Encoding: cfg.Log.Encoding,
		Mode:     cfg.Log.Mode,
	})

	const grpcAddr = "127.0.0.1:50055"
	lnr, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	gatewayFactory := service.NewPaymentGatewayFactory()
	gatewayFactory.RegisterGateway(models.GatewayTypeZalopay, zpGW.NewZalopayGateway(zpGW.ZalopayConfig{
		AppID: cfg.PaymentGateway.Zalopay.AppID,
		Key1:  cfg.PaymentGateway.Zalopay.Key1,
		Key2:  cfg.PaymentGateway.Zalopay.Key2,
		Host:  cfg.PaymentGateway.Zalopay.Host,
	}))

	grpcClients, cleanupGrpc, err := pkgGrpc.InitGrpcClients(cfg.Grpc.OrderSvcAddr, l, cfg.Log.RedactFields)
	if err != nil {
		log.Fatalf("failed to initialize gRPC clients: %v", err)
	}
	defer cleanupGrpc()

	sv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.ValidationInterceptor, interceptors.ErrorHandlerInterceptor),
	)

	paymentSvc := service.NewPaymentService(l, gatewayFactory, grpcClients.Order)

	payment.RegisterPaymentServiceServer(sv, paymentSvc)

	router := mux.NewRouter()
	router.HandleFunc("/zalopay/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var callbackData zpGW.ZalopayCallbackData
		if err := json.NewDecoder(r.Body).Decode(&callbackData); err != nil {
			l.Errorf(r.Context(), "Failed to decode callback data: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if err := service.HandlePaymentCallback(paymentSvc, r.Context(), callbackData, models.GatewayTypeZalopay); err != nil {
			l.Errorf(r.Context(), "Failed to process callback: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}).Methods("POST")

	httpServer := &http.Server{
		Addr:    ":" + cfg.Http.Port,
		Handler: router,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Payment gRPC server started on %s", grpcAddr)
		if err := sv.Serve(lnr); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Start HTTP server
	go func() {
		log.Printf("Payment HTTP server started on port %s", cfg.Http.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to serve HTTP: %v", err)
		}
	}()

	<-sigCh
	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Graceful shutdown for gRPC server
	sv.GracefulStop()
	log.Println("Servers stopped")
}
