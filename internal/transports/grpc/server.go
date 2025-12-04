package grpctransport

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	grpc "google.golang.org/grpc"
)

// ServiceRegistrar is a function that receives a *grpc.Server and registers
// generated gRPC services (pb.RegisterXxxServer). Modules provide these
// functions when wiring the container.
type ServiceRegistrar func(s *grpc.Server)

// Server is a small wrapper around a gRPC server which allows registering
// service registration callbacks and starting/stopping the server.
type Server struct {
	srv        *grpc.Server
	registrars []ServiceRegistrar
	mu         sync.Mutex
}

// NewServer creates a new Server with optional grpc.ServerOptions.
func NewServer(opts ...grpc.ServerOption) *Server {
	return &Server{srv: grpc.NewServer(opts...), registrars: []ServiceRegistrar{}}
}

// RegisterService registers a service registration callback.
// Call this before Start.
func (s *Server) RegisterService(r ServiceRegistrar) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registrars = append(s.registrars, r)
}

// Start binds to the given address (host:port) and runs the server until
// the context is cancelled or an interrupt signal is received.
// It returns any non-nil error from the listener or server.
func (s *Server) Start(ctx context.Context, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// Register services
	s.mu.Lock()
	for _, r := range s.registrars {
		r(s.srv)
	}
	s.mu.Unlock()

	// Run server in goroutine
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- s.srv.Serve(lis)
	}()

	// Watch for context cancellation or OS interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		s.srv.GracefulStop()
		return ctx.Err()
	case sig := <-sigCh:
		_ = sig
		s.srv.GracefulStop()
		return nil
	case err := <-serveErr:
		return err
	}
}

// Stop gracefully stops the server.
func (s *Server) Stop() {
	s.srv.GracefulStop()
}
