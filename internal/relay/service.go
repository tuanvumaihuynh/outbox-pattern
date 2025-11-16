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
	"github.com/tuanvumaihuynh/outbox-pattern/pkg/outbox"
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

				resultChan := make(chan repository.BulkUpdateOutboxMsgsItem, len(outboxMsgs))
				var wg sync.WaitGroup

				for _, outboxMsg := range outboxMsgs {
					msg := outboxMsg
					produceCtx := outbox.ExtractContextFromHeaders(ctx, msg.Headers)
					wg.Go(func() {
						produceMsg := mq.ProduceMsg{
							Topic:        msg.Topic,
							Headers:      msg.Headers,
							Payload:      msg.Payload,
							PartitionKey: msg.PartitionKey,
						}

						if err := s.mqProducer.Produce(produceCtx, produceMsg); err != nil {
							s.logger.ErrorContext(produceCtx,
								"error producing message",
								slog.String("outbox_msg_id", msg.ID.String()),
								slog.String("topic", msg.Topic),
								slog.Any("error", err),
							)
							resultChan <- repository.BulkUpdateOutboxMsgsItem{
								ID:    msg.ID,
								Error: ptr.New(err.Error()),
							}
							return
						}

						resultChan <- repository.BulkUpdateOutboxMsgsItem{
							ID:    msg.ID,
							Error: nil,
						}
					})
				}

				// close channel after all goroutines complete
				go func() {
					wg.Wait()
					close(resultChan)
				}()

				// collect results from channel
				items := make([]repository.BulkUpdateOutboxMsgsItem, 0, len(outboxMsgs))
				for item := range resultChan {
					items = append(items, item)
				}

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
