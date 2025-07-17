package httpserver

import (
	"net"
	"strconv"
	"time"
)

type Option func(*Server)

func Port(port int) Option {
	return func(s *Server) {
		portStr := strconv.Itoa(port)
		s.server.Addr = net.JoinHostPort("", portStr)
	}
}

func ReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.server.ReadTimeout = timeout
	}
}

func WriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.server.WriteTimeout = timeout
	}
}

func ShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = timeout
	}
}
