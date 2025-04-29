package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/vogiaan1904/payment-svc/config"
	"github.com/vogiaan1904/payment-svc/internal/appconfig/mongo"
	"github.com/vogiaan1904/payment-svc/internal/interceptors"
	repository "github.com/vogiaan1904/payment-svc/internal/repositories"
	service "github.com/vogiaan1904/payment-svc/internal/services"
	pkgLog "github.com/vogiaan1904/payment-svc/pkg/log"
	payment "github.com/vogiaan1904/payment-svc/protogen/golang/payment"
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

	// MongoDB connection
	mClient, err := mongo.Connect(cfg.Mongo.DatabaseUri)
	if err != nil {
		panic(err)
	}
	defer mongo.Disconnect(mClient)
	db := mClient.Database(cfg.Mongo.DatabaseName)

	// Repository and Service initialization
	paymentRepo := repository.NewPaymentRepository(l, db)
	paymentSvc := service.NewPaymentService(l, paymentRepo)

	// gRPC server
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.ValidationInterceptor, interceptors.ErrorHandlerInterceptor),
	)

	payment.RegisterPaymentServiceServer(server, paymentSvc)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Payment gRPC server started on %s", addr)
		if err := server.Serve(lnr); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	<-sigCh
	log.Println("Shutting down gRPC server...")

	server.GracefulStop()
	log.Println("Server stopped")
}
