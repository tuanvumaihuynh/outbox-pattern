package apperr

import "github.com/tuanvumaihuynh/outbox-pattern/pkg/zerror"

const (
	ValidationErrorCode = "VALIDATION_FAILED"
)

var (
	ValidationErr = zerror.NewValidationFailed(ValidationErrorCode, "validation error")
)
