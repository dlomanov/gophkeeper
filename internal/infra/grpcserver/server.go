package grpcserver

import (
	"context"
	"google.golang.org/grpc"
	"net"
	"time"
)

const (
	defaultAddr        = ":9090"
	defaultNetwork     = "tcp"
	defaultReadTimeout = 5 * time.Second
)

type (
	Server struct {
		listener        net.Listener
		addr            string
		notify          chan error
		serverOptions   []grpc.ServerOption
		Server          *grpc.Server
		shutdownTimeout time.Duration
	}
)

func New(opts ...Option) *Server {
	var err error

	s := &Server{
		addr:            defaultAddr,
		shutdownTimeout: defaultReadTimeout,
		notify:          make(chan error, 1),
	}
	for _, opt := range opts {
		opt(s)
	}
	s.Server = grpc.NewServer(s.serverOptions...)

	if s.listener == nil {
		s.listener, err = net.Listen(defaultNetwork, s.addr)
		if err != nil {
			s.notify <- err
			return s
		}
	}
	s.start(s.listener)

	return s
}

func (s *Server) start(listener net.Listener) {
	go func() {
		defer close(s.notify)
		s.notify <- s.Server.Serve(listener)
	}()
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		s.Server.GracefulStop()
	}()
	select {
	case <-ctx.Done():
		s.Server.Stop()
	case <-stopped:
		cancel()
	}
	s.listener = nil

	return nil
}
