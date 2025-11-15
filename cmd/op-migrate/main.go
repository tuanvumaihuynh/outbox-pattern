package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/config"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/log"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/storage/db"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("error running migrate application: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	time.Local = time.UTC

	type Config struct {
		Log      config.Log
		Postgres config.Postgres
	}
	cfg, err := config.New[Config]()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	logger := log.NewSlogLogger(cfg.Log)

	pgxPool, err := db.NewPgxPool(ctx, cfg.Postgres)
	if err != nil {
		return fmt.Errorf("error creating pgx pool: %w", err)
	}
	defer pgxPool.Close()

	logger.InfoContext(ctx, "starting database migration")

	if err := db.Migrate(pgxPool); err != nil {
		return fmt.Errorf("error migrating database: %w", err)
	}

	logger.InfoContext(ctx, "database migration completed successfully")

	return nil
}
