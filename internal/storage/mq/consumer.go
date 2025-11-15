package mq

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/config"
)

type HandlerFunc func(ctx context.Context, topic string, payload []byte) error

type CleanupFunc func()

type Consumer interface {
	RegisterHandler(topic string, handler HandlerFunc) error
	Run(ctx context.Context) (CleanupFunc, error)
}

var _ Consumer = (*KafkaConsumer)(nil)

type KafkaConsumer struct {
	cl       *kgo.Client
	handlers map[string]HandlerFunc
	log      *slog.Logger
}

func NewKafkaConsumer(ctx context.Context, cfg config.Kafka, logger *slog.Logger) (*KafkaConsumer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Addresses...),
		kgo.ConsumerGroup(cfg.Group),
		kgo.AllowAutoTopicCreation(),
		kgo.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()
	if err := cl.Ping(pingCtx); err != nil {
		cl.Close()
		return nil, fmt.Errorf("failed to ping kafka: %w", err)
	}

	return &KafkaConsumer{
		cl:       cl,
		handlers: make(map[string]HandlerFunc),
		log:      logger,
	}, nil
}

func (c *KafkaConsumer) RegisterHandler(topic string, handler HandlerFunc) error {
	if _, exists := c.handlers[topic]; exists {
		return fmt.Errorf("handler for topic %s already registered", topic)
	}

	c.cl.AddConsumeTopics(topic)
	c.handlers[topic] = handler
	return nil
}

func (c *KafkaConsumer) Run(ctx context.Context) (CleanupFunc, error) {
	ctx, cancel := context.WithCancel(ctx)
	doneChan := make(chan struct{})
	defer close(doneChan)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				fetches := c.cl.PollFetches(ctx)
				if errs := fetches.Errors(); len(errs) > 0 {
					if errs[0].Err == context.Canceled {
						// context cancelled, likely due to shutdown
						continue
					}

					c.log.ErrorContext(ctx, "error fetching messages",
						slog.Any("error", errs),
					)
					continue
				}

				fetches.EachRecord(func(rec *kgo.Record) {
					defer func() {
						if rvr := recover(); rvr != nil {
							span := trace.SpanFromContext(ctx)
							span.RecordError(fmt.Errorf("panic: %v", rvr))
							span.SetStatus(codes.Error, "panic in handler")

							c.log.ErrorContext(ctx, "panic in message handler",
								slog.String("topic", rec.Topic),
								slog.Any("recover", rvr),
								slog.String("stack", string(debug.Stack())),
							)
						}
					}()

					fn, exists := c.handlers[rec.Topic]
					if !exists {
						c.log.WarnContext(ctx, "no handler registered for topic",
							slog.String("topic", rec.Topic),
						)
						return
					}

					if err := fn(ctx, rec.Topic, rec.Value); err != nil {
						c.log.ErrorContext(ctx, "error handling message",
							slog.String("topic", rec.Topic),
							slog.String("key", string(rec.Key)),
							slog.Any("error", err),
						)
						return
					}
				})

				if err := c.cl.CommitUncommittedOffsets(ctx); err != nil {
					c.log.ErrorContext(ctx, "error committing offsets",
						slog.Any("error", err),
					)
				}
			}
		}
	}()

	cleanup := func() {
		cancel()
		c.cl.Close()
		<-doneChan
	}

	return cleanup, nil
}

func (c *KafkaConsumer) Close() {
	c.cl.Close()
}
