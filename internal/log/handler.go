package log

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"

	"github.com/tuanvumaihuynh/outbox-pattern/pkg/correlationid"
)

var _ slog.Handler = (*enrichedHandler)(nil)

// enrichedHandler enriches logs with trace and correlation data
type enrichedHandler struct {
	h slog.Handler
}

func newEnrichedHandler(h slog.Handler) enrichedHandler {
	return enrichedHandler{h: h}
}

func (eh enrichedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return eh.h.Enabled(ctx, level)
}

func (eh enrichedHandler) Handle(ctx context.Context, r slog.Record) error {
	if correlationID, ok := correlationid.FromContext(ctx); ok {
		r.Add("correlation_id", slog.StringValue(correlationID))
	}

	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if spanCtx.IsValid() {
		r.Add("trace_id", slog.StringValue(spanCtx.TraceID().String()))
		r.Add("span_id", slog.StringValue(spanCtx.SpanID().String()))
	}

	return eh.h.Handle(ctx, r)
}

func (eh enrichedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return newEnrichedHandler(eh.h.WithAttrs(attrs))
}

func (eh enrichedHandler) WithGroup(name string) slog.Handler {
	return newEnrichedHandler(eh.h.WithGroup(name))
}
