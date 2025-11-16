package outbox

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/tuanvumaihuynh/outbox-pattern/pkg/correlationid"
)

// BuildHeaders creates headers map with trace context and correlation ID injected from context.
func BuildHeaders(ctx context.Context) map[string]string {
	headers := map[string]string{}

	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.MapCarrier(headers))

	if correlationID, ok := correlationid.FromContext(ctx); ok {
		headers[correlationid.Header] = correlationID
	}

	return headers
}

// ExtractContextFromHeaders extracts trace context and correlation ID from headers map and injects them into context.
func ExtractContextFromHeaders(ctx context.Context, headers map[string]string) context.Context {
	propagator := otel.GetTextMapPropagator()
	ctx = propagator.Extract(ctx, propagation.MapCarrier(headers))

	if correlationID, ok := headers[correlationid.Header]; ok {
		ctx = correlationid.NewContext(ctx, correlationID)
	}

	return ctx
}

// InjectCorrelationIDFromRecord extracts correlation ID from Kafka record headers and injects it into context.
// Returns the context with correlation ID if found in headers, otherwise returns the original context.
func InjectCorrelationIDFromRecord(ctx context.Context, rec *kgo.Record) context.Context {
	for _, header := range rec.Headers {
		if header.Key == correlationid.Header {
			return correlationid.NewContext(ctx, string(header.Value))
		}
	}
	return ctx
}
