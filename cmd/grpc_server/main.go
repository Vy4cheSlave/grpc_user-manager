package main

import (
	"context"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/config"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/db"
	grpcServer "github.com/Vy4cheSlave/grpc_user-manager/internal/infrastructure/grpc"
	loggerpack "github.com/Vy4cheSlave/grpc_user-manager/internal/logger"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/usecase/auth"
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
	grpcAddr := strings.Join([]string{cfg.GRPC.Host, cfg.GRPC.Port}, ":")

	// Инициализация postgres
	repo, err := db.NewRepository(context.Background(), cfg.PostgreSQL)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error initializing repository"))
	}

	// Инициализация сервиса
	service := auth.NewAuthService(logger, repo, repo, cfg.TokenTTL, cfg.TokenSecret)

	// Инициализация сервера
	server := grpcServer.New(logger, &grpcAddr, service)
	go func() {
		err := server.GRPCServer.Run()
		if err != nil {
			log.Fatal(errors.Wrap(err, "failed to start server"))
		}
	}()

	// Ожидание системных сигналов для корректного завершения работы
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	signalFromChannel := <-signalChan

	logger.Info("Shutting down server...", slog.String("signal", signalFromChannel.String()))
	server.GRPCServer.Stop()
	logger.Info("Shutting down gracefully...")
}
