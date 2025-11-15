package model

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Sku           string    `json:"sku"`
	Price         float64   `json:"price"`
	StockQuantity int       `json:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
