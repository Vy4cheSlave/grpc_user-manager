package grpc

import (
	"context"
	pbAuth "github.com/Vy4cheSlave/grpc_user-manager/gen/go/auth"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/db"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"net"
	"strings"
)

type Auth interface {
	Login(ctx context.Context, username string, password string) (token string, err error)
	Register(ctx context.Context, username string, password string) (userId *string, err error)
}

type Server struct {
	log        *slog.Logger
	service    Auth
	gRPCServer *grpc.Server
	addr       string
}

func NewServer(log *slog.Logger, service Auth, addr *string) *Server {
	gRPCServer := grpc.NewServer()
	RegisterGRPC(gRPCServer, service)
	return &Server{
		log:        log,
		gRPCServer: gRPCServer,
		service:    service,
		addr:       *addr,
	}
}

func (s *Server) Run() error {
	const op = "internal/infrastructure/grpc/server.Server.Run"
	log := s.log.With(slog.String("operation", op), slog.String("addr", s.addr))

	lis, err := net.Listen(
		"tcp",
		s.addr,
	)
	if err != nil {
		return errors.Wrap(err, strings.Join([]string{op, "failed to listen"}, ": "))
	}

	log.Info("grpc server is running")

	if err := s.gRPCServer.Serve(lis); err != nil {
		return errors.Wrap(err, strings.Join([]string{op, "failed to serve grpc server"}, ": "))
	}

	return nil
}

func (s *Server) Stop() {
	const op = "internal/infrastructure/grpc/server.Server.Stop"
	s.log.With(slog.String("operation", op), slog.String("addr", s.addr)).
		Info("grpc server is stopping")
	s.gRPCServer.Stop()
}

type serverAPI struct {
	pbAuth.UnimplementedAuthServer
	auth Auth
}

func RegisterGRPC(gRPC *grpc.Server, auth Auth) {
	pbAuth.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, req *pbAuth.LoginRequest) (*pbAuth.LoginResponse, error) {
	// todo:валидация данных

	token, err := s.auth.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "argument is invalid")
		}
		return nil, status.Errorf(codes.Internal, "%v", err) // todo
	}

	return &pbAuth.LoginResponse{
		Token: token,
	}, nil
}
func (s *serverAPI) Register(ctx context.Context, req *pbAuth.RegisterRequest) (*pbAuth.RegisterResponse, error) {
	// todo:валидация данных

	userId, err := s.auth.Register(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		if errors.Is(err, db.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Errorf(codes.Internal, "%v", err) // todo
	}

	return &pbAuth.RegisterResponse{
		UserId: *userId,
	}, nil
}
