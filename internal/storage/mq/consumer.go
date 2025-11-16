package mq

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/config"
	"github.com/tuanvumaihuynh/outbox-pattern/pkg/outbox"
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
	kTracer  *kotel.Tracer
	handlers map[string]HandlerFunc
	logger   *slog.Logger
}

func NewKafkaConsumer(ctx context.Context, cfg config.Kafka, logger *slog.Logger) (*KafkaConsumer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Addresses...),
		kgo.ConsumerGroup(cfg.Group),
		kgo.AllowAutoTopicCreation(),
		kgo.WithContext(ctx),
		kgo.WithHooks(kTracer),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka client: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()
	if err := cl.Ping(pingCtx); err != nil {
		cl.Close()
		return nil, fmt.Errorf("ping kafka: %w", err)
	}

	return &KafkaConsumer{
		cl:       cl,
		kTracer:  kTracer,
		handlers: make(map[string]HandlerFunc),
		logger:   logger,
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

					c.logger.ErrorContext(ctx, "error fetching messages",
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

							c.logger.ErrorContext(ctx, "panic in message handler",
								slog.String("topic", rec.Topic),
								slog.Any("recover", rvr),
								slog.String("stack", string(debug.Stack())),
							)
						}
					}()

					ctx, span := c.kTracer.WithProcessSpan(rec)
					defer span.End()

					// inject correlation ID from record headers into context
					ctx = outbox.InjectCorrelationIDFromRecord(ctx, rec)

					fn, exists := c.handlers[rec.Topic]
					if !exists {
						span.RecordError(fmt.Errorf("no handler for topic %s", rec.Topic))
						span.SetStatus(codes.Error, "no handler registered for topic")
						c.logger.ErrorContext(ctx, "no handler registered for topic",
							slog.String("topic", rec.Topic),
						)
						return
					}

					if err := fn(ctx, rec.Topic, rec.Value); err != nil {
						span.RecordError(err)
						span.SetStatus(codes.Error, "error in consumer handler")
						c.logger.ErrorContext(ctx, "error handling message",
							slog.String("topic", rec.Topic),
							slog.String("key", string(rec.Key)),
							slog.Any("error", err),
						)
						return
					}

					span.SetStatus(codes.Ok, "")
				})

				if err := c.cl.CommitUncommittedOffsets(ctx); err != nil {
					c.logger.ErrorContext(ctx, "error committing offsets",
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
