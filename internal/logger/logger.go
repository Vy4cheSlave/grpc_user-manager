package logger

import (
	"fmt"
	slogzap "github.com/samber/slog-zap/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log/slog"
)

func NewLogger(level string) (*slog.Logger, error) {
	logLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("error ParseAtomicLevel %s: %w", level, err)
	}

	logger, err := zap.Config{
		Level:       logLevel,
		Encoding:    "json",
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",
			TimeKey:    "ts",
			LevelKey:   "level", //
			EncodeTime: zapcore.RFC3339NanoTimeEncoder,
		},
		DisableStacktrace: true,
	}.Build()
	if err != nil {
		return nil, fmt.Errorf("error logConfig.Build: %w", err)
	}

	handler := slogzap.Option{
		Logger:    logger,
		AddSource: true, // Добавляет информацию об источнике (опционально)
	}.NewZapHandler()

	return slog.New(handler), nil
}
