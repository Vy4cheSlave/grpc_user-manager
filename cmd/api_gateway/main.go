package main

import (
	"github.com/Vy4cheSlave/grpc_user-manager/internal/config"
	grpcClient "github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/getway/client/grpc"
	getwayServer "github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/getway/server"
	loggerpack "github.com/Vy4cheSlave/grpc_user-manager/internal/logger"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	if err := godotenv.Load(config.EnvPath); err != nil {
		log.Fatal("Ошибка загрузки env файла:", err)
	}

	// Загружаем конфигурацию из переменных окружения
	var cfg config.Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(errors.Wrap(err, "failed to load configuration"))
	}

	// Инициализация логгера
	logger, err := loggerpack.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error initializing logger"))
	}

	getwayAddr := strings.Join([]string{cfg.Getway.Host, cfg.Getway.Port}, ":")
	taskCrudAddr := strings.Join([]string{cfg.Rest.Host, cfg.Rest.Port}, ":")
	authAddr := strings.Join([]string{cfg.GRPC.Host, cfg.GRPC.Port}, ":")

	clientConn, err := grpcClient.CreateGrpcClient(authAddr)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer clientConn.Close()
	getwayHandler, err := grpcClient.RegisterAuthServiceHandler(clientConn)
	if err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}
	authClient := grpcClient.NewAuthClient(clientConn)

	//Инициализация сервера
	server := getwayServer.New(
		logger,
		&getwayAddr,
		&cfg.TokenSecret,
		&authAddr,
		&taskCrudAddr,
		getwayHandler,
		authClient,
	)
	go func() {
		err := server.RESTServer.Run()
		if err != nil {
			log.Fatal(errors.Wrap(err, "failed to start server"))
		}
	}()

	// Ожидание системных сигналов для корректного завершения работы
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	signalFromChannel := <-signalChan

	logger.Info("Shutting down server...", slog.String("signal", signalFromChannel.String()))
	logger.Info("Shutting down gracefully...")
}
