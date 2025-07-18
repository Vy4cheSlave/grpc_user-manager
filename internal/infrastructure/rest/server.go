package rest

import (
	"log/slog"
)

type ServerTaskCrud struct {
	RESTServer *Server
}

func New(
	log *slog.Logger,
	addr *string,
	tokenSecret *[]byte,
	service TaskCrud,
) *ServerTaskCrud {
	server := NewServer(log, service, addr, tokenSecret)
	return &ServerTaskCrud{
		RESTServer: server,
	}
}
