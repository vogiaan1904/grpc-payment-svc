package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/vogiaan1904/payment-svc/config"
	"github.com/vogiaan1904/payment-svc/internal/interceptors"
	"github.com/vogiaan1904/payment-svc/internal/models"
	bankTf "github.com/vogiaan1904/payment-svc/internal/services/banktransfer"
	zpGW "github.com/vogiaan1904/payment-svc/internal/services/banktransfer/zalopay"
	pkgGrpc "github.com/vogiaan1904/payment-svc/pkg/grpc"
	pkgLog "github.com/vogiaan1904/payment-svc/pkg/log"
	"github.com/vogiaan1904/payment-svc/protogen/golang/payment"
	"go.temporal.io/sdk/client"
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

	grpcAddr := ":50055"
	lnr, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Temporal client
	tCli, err := client.Dial(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		l.Fatalf(context.Background(), "failed to initialize Temporal client: %v", err)
	}
	defer tCli.Close()
	l.Info(context.Background(), "Temporal Client connected.")

	// gRPC clients
	gprcClis, cleanupGrpc, err := pkgGrpc.InitGrpcClients(cfg.Grpc.OrderSvcAddr, l, cfg.Log.RedactFields)
	if err != nil {
		log.Fatalf("failed to initialize gRPC clients: %v", err)
	}
	defer cleanupGrpc()

	sv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.ValidationInterceptor, interceptors.ErrorHandlerInterceptor),
	)

	// Payment gateways
	zpGW := zpGW.NewZalopayGateway(cfg.PaymentGateway.Zalopay.AppID, cfg.PaymentGateway.Zalopay.Key1, cfg.PaymentGateway.Zalopay.Key2, cfg.PaymentGateway.Zalopay.Host)

	gwf := bankTf.NewPaymentGatewayFactory()
	gwf.RegisterGateway(models.GatewayTypeZalopay, zpGW)

	pmtSvc := bankTf.NewPaymentService(l, gwf, gprcClis.Order, tCli)
	payment.RegisterPaymentServiceServer(sv, pmtSvc)

	go func() {
		l.Info(context.Background(), "Payment gRPC server started on %s", grpcAddr)
		if err := sv.Serve(lnr); err != nil {
			l.Errorf(context.Background(), "failed to serve gRPC: %v", err)
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	l.Info(context.Background(), "Shutting down gRPC server...")

	sv.GracefulStop()
	l.Info(context.Background(), "gRPC server stopped")
}
