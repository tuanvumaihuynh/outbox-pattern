package correlationid

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

var Header = http.CanonicalHeaderKey("X-Correlation-Id")

type ctxKey struct{}

var correlationIDCtxKey = ctxKey{}

// FromContext retrieves the correlation ID from the context.
func FromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(correlationIDCtxKey).(string)
	return val, ok
}

// NewContext creates a new context with the given correlation ID.
func NewContext(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDCtxKey, correlationID)
}

// New creates a new correlation ID.
func New() string {
	return uuid.NewString()
}
