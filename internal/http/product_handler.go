package http

import (
	"context"
	"fmt"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/http/gen"
	"github.com/tuanvumaihuynh/outbox-pattern/internal/service"
)

type productHandler struct {
	productSvc service.ProductService
}

func newProductHandler(productSvc service.ProductService) *productHandler {
	return &productHandler{
		productSvc: productSvc,
	}
}

func (h *productHandler) ListProducts(ctx context.Context, request gen.ListProductsRequestObject) (gen.ListProductsResponseObject, error) {
	products, err := h.productSvc.ListAllProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("product service list all products: %w", err)
	}

	items := make([]gen.ProductResponse, 0, len(products))
	for _, product := range products {
		res := gen.ProductResponse{
			Id:            product.ID,
			Name:          product.Name,
			Sku:           product.Sku,
			Price:         product.Price,
			StockQuantity: product.StockQuantity,
			CreatedAt:     product.CreatedAt,
			UpdatedAt:     product.UpdatedAt,
		}
		items = append(items, res)
	}

	return gen.ListProducts200JSONResponse(items), nil
}

func (h *productHandler) CreateProduct(ctx context.Context, request gen.CreateProductRequestObject) (gen.CreateProductResponseObject, error) {
	params := service.CreateProductParams{
		Name:          request.Body.Name,
		Sku:           request.Body.Sku,
		Price:         request.Body.Price,
		StockQuantity: request.Body.StockQuantity,
	}
	product, err := h.productSvc.CreateProduct(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("product service create product: %w", err)
	}

	res := gen.ProductResponse{
		Id:            product.ID,
		Name:          product.Name,
		Sku:           product.Sku,
		Price:         product.Price,
		StockQuantity: product.StockQuantity,
		CreatedAt:     product.CreatedAt,
		UpdatedAt:     product.UpdatedAt,
	}

	return gen.CreateProduct201JSONResponse(res), nil
}
