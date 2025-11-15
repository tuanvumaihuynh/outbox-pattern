package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/storage/db"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/storage/db/sqlc"
)

type CreateOutboxMsgParams struct {
	Topic        string
	Headers      map[string]string
	Payload      json.RawMessage
	PartitionKey *string
}

type ListUnprocessedOutboxMsgsParams struct {
	BatchSize int32
}

type ListUnprocessedOutboxMsgsResult struct {
	ID           uuid.UUID
	Topic        string
	Headers      map[string]string
	Payload      json.RawMessage
	PartitionKey *string
}

type BulkUpdateOutboxMsgsItem struct {
	ID    uuid.UUID
	Error *string
}

type BulkUpdateOutboxMsgsParams struct {
	Items []BulkUpdateOutboxMsgsItem
}

type OutboxMsgRepository interface {
	WithDB(db db.DB) OutboxMsgRepository
	CreateOutboxMsg(ctx context.Context, params CreateOutboxMsgParams) error
	ListUnprocessedOutboxMsgs(ctx context.Context, params ListUnprocessedOutboxMsgsParams) ([]ListUnprocessedOutboxMsgsResult, error)
	BulkUpdateOutboxMsgs(ctx context.Context, params BulkUpdateOutboxMsgsParams) error
}

type outboxMsgRepository struct {
	db      db.DB
	queries sqlc.Queries
}

func NewOutboxMsgRepository(db db.DB, queries sqlc.Queries) OutboxMsgRepository {
	return &outboxMsgRepository{
		db:      db,
		queries: queries,
	}
}

func (r outboxMsgRepository) WithDB(db db.DB) OutboxMsgRepository {
	return &outboxMsgRepository{
		db:      db,
		queries: r.queries,
	}
}

func (r outboxMsgRepository) CreateOutboxMsg(ctx context.Context, params CreateOutboxMsgParams) error {
	headersBytes, err := json.Marshal(params.Headers)
	if err != nil {
		return fmt.Errorf("marshal headers: %w", err)
	}

	headers := json.RawMessage(headersBytes)
	if err := r.queries.OutboxMsgCreate(ctx, r.db, sqlc.OutboxMsgCreateParams{
		Topic:        params.Topic,
		Headers:      &headers,
		Payload:      params.Payload,
		PartitionKey: params.PartitionKey,
		CreatedAt:    time.Now(),
		ProcessedAt:  nil,
		Error:        nil,
	}); err != nil {
		return fmt.Errorf("outbox msg create: %w", err)
	}

	return nil
}

func (r outboxMsgRepository) ListUnprocessedOutboxMsgs(ctx context.Context, params ListUnprocessedOutboxMsgsParams) ([]ListUnprocessedOutboxMsgsResult, error) {
	msgs, err := r.queries.OutboxMsgListUnprocessed(ctx, r.db, params.BatchSize)
	if err != nil {
		return nil, fmt.Errorf("outbox msg list unprocessed: %w", err)
	}

	results := make([]ListUnprocessedOutboxMsgsResult, 0, len(msgs))
	for _, msg := range msgs {
		headers := map[string]string{}
		if msg.Headers != nil {
			if err := json.Unmarshal(*msg.Headers, &headers); err != nil {
				return nil, fmt.Errorf("unmarshal headers: %w", err)
			}
		}

		results = append(results, ListUnprocessedOutboxMsgsResult{
			ID:           msg.ID,
			Topic:        msg.Topic,
			Headers:      headers,
			Payload:      msg.Payload,
			PartitionKey: msg.PartitionKey,
		})
	}

	return results, nil
}

func (r outboxMsgRepository) BulkUpdateOutboxMsgs(ctx context.Context, params BulkUpdateOutboxMsgsParams) error {
	ids := make([]uuid.UUID, 0, len(params.Items))
	errs := make([]*string, 0, len(params.Items))
	for _, item := range params.Items {
		ids = append(ids, item.ID)
		if item.Error != nil {
			errs = append(errs, item.Error)
		} else {
			errs = append(errs, nil)
		}
	}

	_, err := r.db.Exec(ctx, `
		UPDATE outbox_messages AS o
		SET
			processed_at = NOW(),
			error        = e.error
		FROM (
			SELECT
				id,
				error
			FROM (
				SELECT UNNEST(@ids::uuid[])  AS id,
					UNNEST(@errors::text[]) AS error
			) AS t
		) AS e
		WHERE o.id = e.id;
	`, pgx.NamedArgs{
		"ids":    ids,
		"errors": errs,
	})
	if err != nil {
		return fmt.Errorf("outbox msg bulk update: %w", err)
	}

	return nil
}
