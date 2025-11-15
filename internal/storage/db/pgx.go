package db

import (
	"context"
	"fmt"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/config"
)

// NewPgxPool creates a new pgx pool with the given configuration.
func NewPgxPool(ctx context.Context, cfg config.Postgres) (*pgxpool.Pool, error) {
	pgConf, err := pgxpool.ParseConfig(connectionString(cfg))
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	pgConf.ConnConfig.Tracer = newTracer()

	pgConf.MaxConns = cfg.MaxConns
	pgConf.MinConns = cfg.MinConns
	pgConf.MaxConnLifetime = cfg.MaxConnLifetime
	pgConf.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, pgConf)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := otelpgx.RecordStats(pool); err != nil {
		return nil, fmt.Errorf("record database stats: %w", err)
	}

	// Create a context with timeout for ping
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()

	if err := pool.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

func connectionString(cfg config.Postgres) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB, cfg.SSLMode)
}
