package config

import "time"

type Postgres struct {
	Host     string `env:"POSTGRES_HOST,required"`
	Port     int    `env:"POSTGRES_PORT,required"`
	User     string `env:"POSTGRES_USER,required"`
	Password string `env:"POSTGRES_PASSWORD,required"`
	DB       string `env:"POSTGRES_DB,required"`
	SSLMode  string `env:"POSTGRES_SSL_MODE,required"`

	MaxConns        int32         `env:"POSTGRES_MAX_CONNS,required"`
	MinConns        int32         `env:"POSTGRES_MIN_CONNS,required"`
	MaxConnLifetime time.Duration `env:"POSTGRES_MAX_CONN_LIFETIME,required"`
	MaxConnIdleTime time.Duration `env:"POSTGRES_MAX_CONN_IDLE_TIME,required"`
}
