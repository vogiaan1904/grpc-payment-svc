package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

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

	// Initialize logger
	l := pkgLog.InitializeZapLogger(pkgLog.ZapConfig{
		Level:    cfg.Log.Level,
		Encoding: cfg.Log.Encoding,
		Mode:     cfg.Log.Mode,
	})

	const addr = "127.0.0.1:50055"
	lnr, err := net.Listen("tcp", addr)
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

	// Create payment service with the factory
	paymentSvc := service.NewPaymentService(l, gatewayFactory, grpcClients.Order)

	// gRPC server
	sv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.ValidationInterceptor, interceptors.ErrorHandlerInterceptor),
	)

	// Register payment service with gRPC server
	payment.RegisterPaymentServiceServer(sv, paymentSvc)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Payment gRPC server started on %s", addr)
		if err := sv.Serve(lnr); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	<-sigCh
	log.Println("Shutting down gRPC server...")

	sv.GracefulStop()
	log.Println("Server stopped")
}
