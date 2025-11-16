-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_outbox_messages_unprocessed_created_at_asc
ON outbox_messages (created_at ASC)
WHERE processed_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_outbox_messages_unprocessed_created_at_asc;
-- +goose StatementEnd
