package grpcserver

import (
	"crypto/tls"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"time"
)

const (
	ListenerKey ContextKey = "grpc_listener"
)

type (
	Option     func(*Server)
	ContextKey string
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

func TLSCert(cert, certKey []byte) Option {
	if len(cert) == 0 || len(certKey) == 0 {
		panic("cert should be specified")
	}
	crt, err := tls.X509KeyPair(cert, certKey)
	if err != nil {
		panic(fmt.Errorf("failed to load TLS-cert: %w", err))
	}
	return func(s *Server) {
		creds := credentials.NewServerTLSFromCert(&crt)
		s.serverOptions = append(s.serverOptions, grpc.Creds(creds))
	}
}

func ServerOptions(opts ...grpc.ServerOption) Option {
	return func(s *Server) {
		s.serverOptions = append(s.serverOptions, opts...)
	}
}
