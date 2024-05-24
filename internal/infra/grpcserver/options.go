package grpcserver

import (
	"google.golang.org/grpc"
	"time"
)

type Option func(*Server)

func Addr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

func ShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = timeout
	}
}

func ServerOptions(opts ...grpc.ServerOption) Option {
	return func(s *Server) {
		s.serverOptions = opts
	}
}
