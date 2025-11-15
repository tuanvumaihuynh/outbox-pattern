package repository

import (
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/model"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/storage/db"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/storage/db/sqlc"
)

type ProductRepository interface {
	WithDB(db db.DB) ProductRepository
	CreateProduct(ctx context.Context, product model.Product) error
	ListAllProducts(ctx context.Context) ([]model.Product, error)
}

type productRepository struct {
	db      db.DB
	queries sqlc.Queries
}

func NewProductRepository(db db.DB, queries sqlc.Queries) ProductRepository {
	return &productRepository{
		db:      db,
		queries: queries,
	}
}

func (r productRepository) WithDB(db db.DB) ProductRepository {
	return &productRepository{
		db:      db,
		queries: r.queries,
	}
}

func (r productRepository) CreateProduct(ctx context.Context, product model.Product) error {
	var price pgtype.Numeric
	if err := price.Scan(fmt.Sprintf("%f", product.Price)); err != nil {
		return fmt.Errorf("scan price: %w", err)
	}

	if product.StockQuantity > math.MaxInt32 || product.StockQuantity < math.MinInt32 {
		return fmt.Errorf("stock quantity out of range: %d", product.StockQuantity)
	}

	if err := r.queries.ProductCreate(ctx, r.db, sqlc.ProductCreateParams{
		ID:            product.ID,
		Name:          product.Name,
		Sku:           product.Sku,
		Price:         price,
		StockQuantity: int32(product.StockQuantity),
		CreatedAt:     product.CreatedAt,
		UpdatedAt:     product.UpdatedAt,
	}); err != nil {
		return fmt.Errorf("create product: %w", err)
	}

	return nil
}

func (r productRepository) ListAllProducts(ctx context.Context) ([]model.Product, error) {
	products, err := r.queries.ProductListAll(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("list all products: %w", err)
	}

	modelProducts := make([]model.Product, 0, len(products))
	for _, product := range products {
		modelProduct, err := sqlcProductToModelProduct(product)
		if err != nil {
			return nil, fmt.Errorf("convert product to model product: %w", err)
		}
		modelProducts = append(modelProducts, modelProduct)
	}

	return modelProducts, nil
}

func sqlcProductToModelProduct(product sqlc.Product) (model.Product, error) {
	price, err := product.Price.Float64Value()
	if err != nil {
		return model.Product{}, fmt.Errorf("convert price to float64: %w", err)
	}

	return model.Product{
		ID:            product.ID,
		Name:          product.Name,
		Sku:           product.Sku,
		Price:         price.Float64,
		StockQuantity: int(product.StockQuantity),
		CreatedAt:     product.CreatedAt,
		UpdatedAt:     product.UpdatedAt,
	}, nil
}
