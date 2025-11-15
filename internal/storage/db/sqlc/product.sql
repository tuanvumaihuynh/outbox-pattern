-- name: ProductCreate :exec
INSERT INTO products (
	id,
	name,
	sku,
	price,
	stock_quantity,
	created_at,
	updated_at
) VALUES (
	@id,
	@name,
	@sku,
	@price,
	@stock_quantity,
	@created_at,
	@updated_at
);

-- name: ProductListAll :many
SELECT * FROM products;
