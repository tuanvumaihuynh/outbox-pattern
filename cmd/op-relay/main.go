package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/config"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/log"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/relay"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/repository"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/storage/db"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/storage/db/sqlc"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/storage/mq"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/telemetry"
	"github.com/tuanvumaihuynh/outbox-pattern/pkg/cmdutil"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("error running relay application: %v\n", err)
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
		Relay    config.Relay
		Kafka    config.Kafka
		Otel     config.Otel
	}
	cfg, err := config.New[Config]()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	logger := log.NewSlogLogger(cfg.Log)

	cleanupTracer, err := telemetry.InitTracer(ctx, cfg.Otel)
	if err != nil {
		return fmt.Errorf("error initializing tracer: %w", err)
	}
	defer func() {
		if err := cleanupTracer(ctx); err != nil {
			logger.ErrorContext(ctx, "error cleaning up tracer", slog.Any("error", err))
		}
	}()

	pgxPool, err := db.NewPgxPool(ctx, cfg.Postgres)
	if err != nil {
		return fmt.Errorf("error creating pgx pool: %w", err)
	}
	defer pgxPool.Close()

	dbClient := db.NewClient(pgxPool)
	queries := *sqlc.New()

	kafkaProducer, err := mq.NewKafkaProducer(ctx, cfg.Kafka)
	if err != nil {
		return fmt.Errorf("error creating kafka producer: %w", err)
	}
	defer kafkaProducer.Close()

	outboxMsgRepository := repository.NewOutboxMsgRepository(dbClient, queries)

	interruptChan := cmdutil.InterruptChan()

	svc := relay.NewService(cfg.Relay, logger, dbClient, outboxMsgRepository, kafkaProducer)
	cleanup := svc.Run(ctx)
	logger.InfoContext(ctx, "relay service started")

	<-interruptChan

	logger.InfoContext(ctx, "relay service is shutting down")
	cleanup()

	logger.InfoContext(ctx, "relay service is stopped")

	return nil
}
