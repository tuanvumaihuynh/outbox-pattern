package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/storage/mq"
)

// Service is the event service.
type Service struct {
	logger     *slog.Logger
	mqConsumer mq.Consumer
}

// New creates a new event service.
func New(
	logger *slog.Logger,
	mqConsumer mq.Consumer,
) *Service {
	return &Service{
		logger:     logger,
		mqConsumer: mqConsumer,
	}
}

type CleanupFunc func()

func (s *Service) Run(ctx context.Context) (CleanupFunc, error) {
	if err := s.mqConsumer.RegisterHandler(
		TopicProductCreated,
		func(ctx context.Context, topic string, payload []byte) error {
			var ev ProductCreatedEvent
			if err := json.Unmarshal(payload, &ev); err != nil {
				return fmt.Errorf("unmarshal product created event: %w", err)
			}

			if err := s.handleProductCreatedEvent(ctx, ev); err != nil {
				return fmt.Errorf("handle product created event: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("register product created event handler: %w", err)
	}

	mqCleanup, err := s.mqConsumer.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("run mq consumer: %w", err)
	}

	cleanup := func() {
		mqCleanup()
	}

	return cleanup, nil
}
