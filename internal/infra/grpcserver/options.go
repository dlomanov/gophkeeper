package grpcserver

import (
	"google.golang.org/grpc"
	"net"
	"time"
)

const (
	ListenerKey OptionKey = "grpc_listener"
)

type (
	Option    func(*Server)
	OptionKey string
)

func Listener(l net.Listener) Option {
	return func(s *Server) {
		s.listener = l
	}
}

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
