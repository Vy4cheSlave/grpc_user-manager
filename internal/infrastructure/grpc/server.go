package grpc

import (
	"log/slog"
)

type ServerAuth struct {
	GRPCServer *Server
}

func New(
	log *slog.Logger,
	grpcAddr *string,
	//tokenTTL time.Duration,
	service Auth,
) *ServerAuth {
	server := NewServer(log, service, grpcAddr)
	return &ServerAuth{
		GRPCServer: server,
	}
}
