package event

import (
	"context"
	"log/slog"
)

const TopicProductCreated = "product.created"

type ProductCreatedEvent struct {
	ProductID     string  `json:"product_id"`
	Name          string  `json:"name"`
	Sku           string  `json:"sku"`
	Price         float64 `json:"price"`
	StockQuantity int     `json:"stock_quantity"`
}

func (s *Service) handleProductCreatedEvent(ctx context.Context, ev ProductCreatedEvent) error {
	s.logger.InfoContext(ctx, "handling product created event", slog.Any("event", ev))
	return nil
}
