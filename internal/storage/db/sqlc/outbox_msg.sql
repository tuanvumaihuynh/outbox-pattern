-- name: OutboxMsgCreate :exec
INSERT INTO outbox_messages (
	topic,
	headers,
	payload,
	partition_key,
	created_at,
	processed_at,
	error
) VALUES (
	@topic,
	@headers,
	@payload,
	@partition_key,
	@created_at,
	@processed_at,
	@error
);

-- name: OutboxMsgListUnprocessed :many
SELECT
	id,
	topic,
	headers,
	payload,
	partition_key
FROM outbox_messages
WHERE processed_at IS NULL
ORDER BY created_at ASC
LIMIT @batchSize
FOR UPDATE SKIP LOCKED;
