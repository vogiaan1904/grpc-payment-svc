package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vogiaan1904/payment-svc/config"
	"github.com/vogiaan1904/payment-svc/internal/httpserver"
	"github.com/vogiaan1904/payment-svc/internal/models"
	service "github.com/vogiaan1904/payment-svc/internal/services"
	zpGW "github.com/vogiaan1904/payment-svc/internal/services/zalopay"
	pkgGrpc "github.com/vogiaan1904/payment-svc/pkg/grpc"
	pkgLog "github.com/vogiaan1904/payment-svc/pkg/log"
	"go.temporal.io/sdk/client"
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

	// Initialize Temporal client
	tCli, err := client.Dial(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		l.Fatalf(context.Background(), "failed to initialize Temporal client: %v", err)
	}
	defer tCli.Close()
	l.Info(context.Background(), "Temporal Client connected.")

	// Initialize payment gateways
	gatewayFactory := service.NewPaymentGatewayFactory()
	gatewayFactory.RegisterGateway(models.GatewayTypeZalopay, zpGW.NewZalopayGateway(zpGW.ZalopayConfig{
		AppID: cfg.PaymentGateway.Zalopay.AppID,
		Key1:  cfg.PaymentGateway.Zalopay.Key1,
		Key2:  cfg.PaymentGateway.Zalopay.Key2,
		Host:  cfg.PaymentGateway.Zalopay.Host,
	}))

	// Initialize gRPC clients
	grpcClients, cleanupGrpc, err := pkgGrpc.InitGrpcClients(cfg.Grpc.OrderSvcAddr, l, cfg.Log.RedactFields)
	if err != nil {
		l.Fatalf(context.Background(), "failed to initialize gRPC clients: %v", err)
	}
	defer cleanupGrpc()

	paymentSvc := service.NewPaymentService(l, gatewayFactory, grpcClients.Order, tCli)

	httpAddr := ":" + cfg.Http.Port
	httpServer := httpserver.New(httpAddr, l, paymentSvc)

	go func() {
		if err := httpServer.Start(); err != nil {
			l.Fatalf(context.Background(), "HTTP server error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	l.Info(context.Background(), "Shutting down HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		l.Errorf(context.Background(), "HTTP server shutdown error: %v", err)
	}

	l.Info(context.Background(), "HTTP server stopped")
}
