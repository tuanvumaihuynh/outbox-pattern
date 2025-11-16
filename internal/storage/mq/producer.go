package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/config"
)

type ProduceMsg struct {
	Topic        string
	Headers      map[string]string
	Payload      []byte
	PartitionKey *string
}

type Producer interface {
	Produce(ctx context.Context, msg ProduceMsg) error
}

var (
	_ Producer = (*KafkaProducer)(nil)
)

type KafkaProducer struct {
	cl *kgo.Client
}

func NewKafkaProducer(ctx context.Context, cfg config.Kafka) (*KafkaProducer, error) {
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

	return &KafkaProducer{cl: cl}, nil
}

func (p *KafkaProducer) Produce(ctx context.Context, msg ProduceMsg) error {
	ctx, span := tracer.Start(ctx, "KafkaProducer.Produce",
		trace.WithAttributes(
			attribute.String("topic", msg.Topic),
		),
	)
	defer span.End()

	var msgErr error
	doneChan := make(chan struct{})
	promise := func(r *kgo.Record, err error) {
		msgErr = err
		close(doneChan)
	}

	record := buildProduceRecord(msg)
	p.cl.Produce(ctx, record, promise)

	waitForProduce := func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-doneChan:
			return msgErr
		}
	}

	if err := waitForProduce(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to produce message")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (p *KafkaProducer) Close() {
	p.cl.Close()
}

func buildProduceRecord(msg ProduceMsg) *kgo.Record {
	headers := make([]kgo.RecordHeader, 0, len(msg.Headers))
	for k, v := range msg.Headers {
		headers = append(headers, kgo.RecordHeader{
			Key:   k,
			Value: []byte(v),
		})
	}

	r := &kgo.Record{
		Topic:   msg.Topic,
		Value:   msg.Payload,
		Headers: headers,
	}

	if msg.PartitionKey != nil {
		r.Key = []byte(*msg.PartitionKey)
	}

	return r
}
