INSERT INTO outbox_messages (id, topic, headers, payload, partition_key, created_at)
SELECT
    uuidv7(),
    'product.created' AS topic,
    NULL::jsonb AS headers,
    jsonb_build_object(
        'product_id', uuidv7()::text,
        'name', 'Product ' || gs,
        'sku', 'SKU-' || lpad(gs::text, 5, '0'),
        'price', round((random() * 900 + 100)::numeric, 2),
        'stock_quantity', (random() * 200)::int
    ) AS payload,
    ('product-' || gs) AS partition_key,
    NOW() - (random() * interval '5 days') AS created_at
FROM generate_series(1, 1000000) AS gs;
