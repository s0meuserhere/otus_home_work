package internalhttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/server/http/gen"
	eventservice "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/service/event"
)

var _ gen.ServerInterface = (*Server)(nil)

type Server struct {
	log        *logger.Logger
	events     eventservice.Service
	httpServer *http.Server
}

func NewServer(log *logger.Logger, addr string, events eventservice.Service) *Server {
	s := &Server{
		log:    log,
		events: events,
	}

	mux := http.NewServeMux()
	handler := gen.HandlerFromMux(s, mux)
	mux.HandleFunc("GET /openapi.json", s.serveOpenAPISpec)

	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           loggingMiddleware(log, handler),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return s
}

func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}

func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	s.log.Info("http server started", "addr", s.httpServer.Addr)

	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("serve: %w", err)
		}

		return nil
	}
}

func (s *Server) Stop(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	s.log.Info("http server stopped")

	return nil
}
