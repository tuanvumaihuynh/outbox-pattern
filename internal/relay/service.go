package relay

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/config"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/repository"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/storage/db"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/storage/mq"
	"github.com/tuanvumaihuynh/outbox-pattern/pkg/ptr"
)

type Service struct {
	cfg           config.Relay
	logger        *slog.Logger
	db            db.DB
	outboxMsgRepo repository.OutboxMsgRepository
	mqProducer    mq.Producer

	stopChan chan struct{}
}

func NewService(
	cfg config.Relay,
	logger *slog.Logger,
	db db.DB,
	outboxMsgRepo repository.OutboxMsgRepository,
	mqProducer mq.Producer,
) *Service {
	return &Service{
		cfg:           cfg,
		logger:        logger.With(slog.String("service", "relay")),
		db:            db,
		outboxMsgRepo: outboxMsgRepo,
		mqProducer:    mqProducer,
		stopChan:      make(chan struct{}),
	}
}

type CleanupFunc func()

func (s *Service) Run(ctx context.Context) CleanupFunc {
	ctx, cancel := context.WithCancel(ctx)

	stoppedChan := make(chan struct{})
	go func() {
		defer close(stoppedChan)
		s.run(ctx)
	}()

	return func() {
		close(s.stopChan)
		select {
		case <-stoppedChan:
		case <-time.After(5 * time.Second):
			cancel()
		}
	}
}

func (s *Service) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-time.After(s.cfg.Interval):
			if err := s.db.WithTx(ctx, func(db db.DB) error {
				outboxMsgs, err := s.outboxMsgRepo.
					WithDB(db).
					ListUnprocessedOutboxMsgs(ctx, repository.ListUnprocessedOutboxMsgsParams{
						//nolint:gosec
						BatchSize: int32(s.cfg.BatchSize),
					})
				if err != nil {
					return fmt.Errorf("list unprocessed outbox msgs: %w", err)
				}

				if len(outboxMsgs) == 0 {
					return nil
				}

				s.logger.InfoContext(ctx, "relaying outbox msgs", slog.Int("count", len(outboxMsgs)))

				items := make([]repository.BulkUpdateOutboxMsgsItem, 0, len(outboxMsgs))
				var (
					mu sync.Mutex
					wg sync.WaitGroup
				)

				for _, outboxMsg := range outboxMsgs {
					msg := outboxMsg
					wg.Go(func() {
						produceFunc := func() error {
							produceMsg := mq.ProduceMsg{
								Topic:        msg.Topic,
								Headers:      msg.Headers,
								Payload:      msg.Payload,
								PartitionKey: msg.PartitionKey,
							}
							if err := s.mqProducer.Produce(ctx, produceMsg); err != nil {
								return fmt.Errorf("produce message: %w", err)
							}

							return nil
						}

						if err := produceFunc(); err != nil {
							s.logger.ErrorContext(ctx,
								"error producing message",
								slog.String("outbox_msg_id", msg.ID.String()),
								slog.String("topic", msg.Topic),
								slog.Any("error", err),
							)
							mu.Lock()
							items = append(items, repository.BulkUpdateOutboxMsgsItem{
								ID:    msg.ID,
								Error: ptr.New(err.Error()),
							})
							mu.Unlock()
							return
						}

						mu.Lock()
						items = append(items, repository.BulkUpdateOutboxMsgsItem{
							ID:    msg.ID,
							Error: nil,
						})
						mu.Unlock()
					})
				}

				wg.Wait()

				if err := s.outboxMsgRepo.
					WithDB(db).
					BulkUpdateOutboxMsgs(ctx, repository.BulkUpdateOutboxMsgsParams{
						Items: items,
					}); err != nil {
					return fmt.Errorf("bulk update outbox msgs: %w", err)
				}

				return nil
			}); err != nil {
				s.logger.ErrorContext(ctx, "error relaying outbox msgs", slog.Any("error", err))
				continue
			}
		}
	}
}
