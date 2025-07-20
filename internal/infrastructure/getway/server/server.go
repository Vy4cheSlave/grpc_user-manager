package server

import (
	"github.com/Vy4cheSlave/grpc_user-manager/gen/go/auth"
	"log/slog"
	"net/http"
)

type ServerGateway struct {
	RESTServer *Server
}

func New(
	log *slog.Logger,
	addr *string,
	tokenSecret *[]byte,
	authServiceAddr *string,
	crudServiceAddr *string,
	gatewayHandler http.Handler,
	authClient auth.AuthClient,
) *ServerGateway {
	server := NewServer(
		log,
		addr,
		tokenSecret,
		authServiceAddr,
		crudServiceAddr,
		gatewayHandler,
		authClient,
	)
	return &ServerGateway{
		RESTServer: server,
	}
}
