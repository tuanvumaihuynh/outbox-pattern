-- +goose Up
-- +goose StatementBegin
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name            TEXT NOT NULL,
    sku             TEXT UNIQUE NOT NULL,
    price           NUMERIC(10,2) NOT NULL,
    stock_quantity  INTEGER NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE outbox_messages (
	id              UUID PRIMARY KEY DEFAULT uuidv7(),
	topic           TEXT NOT NULL,
	headers         JSONB,
	payload         JSONB NOT NULL,
	partition_key   TEXT,
	created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	processed_at    TIMESTAMPTZ,
	error 			TEXT
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE outbox_messages;
DROP TABLE products;
-- +goose StatementEnd
