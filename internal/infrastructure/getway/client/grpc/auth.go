package grpc

import (
	"context"
	gw "github.com/Vy4cheSlave/grpc_user-manager/gen/go/auth"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func CreateGrpcClient(addr string) (*grpc.ClientConn, error) {
	grpcClient, err := grpc.NewClient(
		addr, // Адрес gRPC-сервера
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return grpcClient, nil
}

func RegisterAuthServiceHandler(grpcClient *grpc.ClientConn) (*runtime.ServeMux, error) {
	mux := runtime.NewServeMux()
	err := gw.RegisterAuthHandler(context.Background(), mux, grpcClient)
	if err != nil {
		return nil, err
	}
	return mux, nil
}

func NewAuthClient(clientConnInterface grpc.ClientConnInterface) gw.AuthClient {
	return gw.NewAuthClient(clientConnInterface)
}
