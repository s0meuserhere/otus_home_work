package internalgrpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	"github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/server/grpc/gen"
	eventservice "github.com/s0meuserhere/otus_home_work/hw12_13_14_15_calendar/internal/service/event"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var _ gen.EventServiceServer = (*Server)(nil)

type Server struct {
	gen.UnimplementedEventServiceServer
	log        *logger.Logger
	events     eventservice.Service
	grpcServer *grpc.Server
	addr       string
}

func NewServer(log *logger.Logger, addr string, events eventservice.Service) *Server {
	s := &Server{
		log:    log,
		events: events,
		addr:   addr,
	}

	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(s.loggingInterceptor),
	)
	gen.RegisterEventServiceServer(s.grpcServer, s)
	reflection.Register(s.grpcServer)

	return s
}

func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	s.log.Info("grpc server started", "addr", s.addr)

	return s.serve(ctx, ln)
}

func (s *Server) ServeListener(ctx context.Context, ln net.Listener) error {
	return s.serve(ctx, ln)
}

func (s *Server) serve(ctx context.Context, ln net.Listener) error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.grpcServer.Serve(ln); err != nil {
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

func (s *Server) Stop(_ context.Context) error {
	s.grpcServer.GracefulStop()
	s.log.Info("grpc server stopped")

	return nil
}

func (s *Server) loggingInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()
	ctx = logger.WithContext(ctx, s.log.Slog())

	resp, err := handler(ctx, req)

	peerAddr := ""
	if p, ok := peer.FromContext(ctx); ok {
		peerAddr = p.Addr.String()
	}

	code := codes.OK
	if err != nil {
		code = status.Code(err)
	}

	s.log.Info(fmt.Sprintf(
		`%s [%s] %s %s %d`,
		peerAddr,
		start.Format("02/Jan/2006:15:04:05 -0700"),
		info.FullMethod,
		code.String(),
		time.Since(start).Milliseconds(),
	))

	return resp, err
}
