package apperr

import "github.com/tuanvumaihuynh/outbox-pattern/pkg/zerror"

const (
	ValidationFailed        = "VALIDATION_FAILED"
	ProduceSkuAlreadyExists = "PRODUCE_SKU_ALREADY_EXISTS"
)

var (
	ValidationErr              = zerror.NewValidationFailed(ValidationFailed, "validation error")
	ProduceSkuAlreadyExistsErr = zerror.NewConflict(ProduceSkuAlreadyExists, "sku already exists")
)
