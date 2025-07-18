package config

import "time"

const (
	EnvPath = ".env"
)

type Config struct {
	GRPC        GRPC
	PostgreSQL  PostgreSQL
	TokenTTL    time.Duration `envconfig:"TOKEN_TTL_H" required:"true"`
	TokenSecret []byte        `envconfig:"TOKEN_SECRET" required:"true"`
	LogLevel    string        `envconfig:"LOG_LEVEL" required:"true"`
	Rest        Rest
}

type Rest struct {
	Host string `envconfig:"REST_HOST" default:"localhost"`
	Port string `envconfig:"REST_PORT" required:"true"`
}

type GRPC struct {
	Host string `envconfig:"GRPC_HOST" default:"localhost"`
	Port string `envconfig:"GRPC_PORT" required:"true"`
}

type PostgreSQL struct {
	Host                string        `envconfig:"DB_HOST" required:"true"`
	Port                int           `envconfig:"DB_PORT" required:"true"`
	Name                string        `envconfig:"DB_NAME" required:"true"`
	User                string        `envconfig:"DB_USER" required:"true"`
	Password            string        `envconfig:"DB_PASSWORD" required:"true"`
	SSLMode             string        `envconfig:"DB_SSL_MODE" default:"disable"`
	PoolMaxConns        int           `envconfig:"DB_POOL_MAX_CONNS" default:"5"`
	PoolMaxConnLifetime time.Duration `envconfig:"DB_POOL_MAX_CONN_LIFETIME" default:"180s"`
	PoolMaxConnIdleTime time.Duration `envconfig:"DB_POOL_MAX_CONN_IDLE_TIME" default:"100s"`
}
