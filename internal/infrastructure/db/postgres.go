package db

import (
	"context"
	"fmt"
	"github.com/Vy4cheSlave/grpc_user-manager/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

// NewRepository - создание нового экземпляра репозитория с подключением к PostgreSQL
func NewRepository(ctx context.Context, cfg config.PostgreSQL) (*Repository, error) {
	// Формируем строку подключения
	connString := fmt.Sprintf(
		`user=%s password=%s host=%s port=%d dbname=%s sslmode=%s 
        pool_max_conns=%d pool_max_conn_lifetime=%s pool_max_conn_idle_time=%s`,
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
		cfg.PoolMaxConns,
		cfg.PoolMaxConnLifetime.String(),
		cfg.PoolMaxConnIdleTime.String(),
	)

	// Парсим конфигурацию подключения
	parsedConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse PostgreSQL config")
	}

	// Оптимизация выполнения запросов (кеширование запросов)
	parsedConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	// Создаём пул соединений с базой данных
	pool, err := pgxpool.NewWithConfig(ctx, parsedConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create PostgreSQL connection pool")
	}

	return &Repository{pool}, nil
}
