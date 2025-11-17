package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/apperr"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/event"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/model"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/repository"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/storage/db"
	"github.com/tuanvumaihuynh/outbox-pattern/pkg/outbox"
)

type CreateProductParams struct {
	Name          string
	Sku           string
	Price         float64
	StockQuantity int
}

type ProductService interface {
	CreateProduct(ctx context.Context, params CreateProductParams) (model.Product, error)
	ListAllProducts(ctx context.Context) ([]model.Product, error)
}

type productService struct {
	db            db.DB
	productRepo   repository.ProductRepository
	outboxMsgRepo repository.OutboxMsgRepository
}

func NewProductService(
	db db.DB,
	productRepo repository.ProductRepository,
	outboxMsgRepo repository.OutboxMsgRepository,
) ProductService {
	return &productService{
		db:            db,
		productRepo:   productRepo,
		outboxMsgRepo: outboxMsgRepo,
	}
}

func (s *productService) CreateProduct(ctx context.Context, params CreateProductParams) (model.Product, error) {
	ctx, span := tracer.Start(ctx, "productService.CreateProduct")
	defer span.End()

	id, err := uuid.NewV7()
	if err != nil {
		return model.Product{}, fmt.Errorf("generate uuid v7: %w", err)
	}

	now := time.Now()
	product := model.Product{
		ID:            id,
		Name:          params.Name,
		Sku:           params.Sku,
		Price:         params.Price,
		StockQuantity: params.StockQuantity,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	ev := event.ProductCreatedEvent{
		ProductID:     product.ID.String(),
		Name:          product.Name,
		Sku:           product.Sku,
		Price:         product.Price,
		StockQuantity: product.StockQuantity,
	}

	evBytes, err := json.Marshal(ev)
	if err != nil {
		return model.Product{}, fmt.Errorf("marshal event: %w", err)
	}

	if err := s.db.WithTx(ctx, func(dbtx db.DB) error {
		if err := s.productRepo.
			WithDB(dbtx).
			CreateProduct(ctx, product); err != nil {
			if db.IsUniqueViolationError(err, "sku") {
				return apperr.ProduceSkuAlreadyExistsErr.WrapParent(err)
			}
			return fmt.Errorf("product repository create product: %w", err)
		}

		if err := s.outboxMsgRepo.
			WithDB(dbtx).
			CreateOutboxMsg(ctx, repository.CreateOutboxMsgParams{
				Topic:   event.TopicProductCreated,
				Headers: outbox.BuildHeaders(ctx),
				Payload: evBytes,
			}); err != nil {
			return fmt.Errorf("outbox msg repository create outbox msg: %w", err)
		}

		return nil
	}); err != nil {
		return model.Product{}, fmt.Errorf("db with tx: %w", err)
	}

	return product, nil
}

func (s *productService) ListAllProducts(ctx context.Context) ([]model.Product, error) {
	products, err := s.productRepo.ListAllProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("product repository list all products: %w", err)
	}

	return products, nil
}
