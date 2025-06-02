package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vogiaan1904/payment-svc/pkg/log"
	"github.com/vogiaan1904/payment-svc/protogen/golang/payment"
)

type Server struct {
	server     *http.Server
	logger     log.Logger
	paymentSvc payment.PaymentServiceServer
}

func New(addr string, logger log.Logger, paymentSvc payment.PaymentServiceServer) *Server {
	router := mux.NewRouter()

	server := &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
		logger:     logger,
		paymentSvc: paymentSvc,
	}

	server.registerRoutes(router)

	return server
}

func (s *Server) Start() error {
	s.logger.Info(context.Background(), "Starting HTTP server on %s", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info(ctx, "Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

func (s *Server) registerRoutes(router *mux.Router) {
	router.HandleFunc("/zalopay/callback", s.handleZalopayCallback).Methods(http.MethodPost)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods(http.MethodGet)
}
