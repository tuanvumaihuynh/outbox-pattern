package middleware

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"
	"go.opentelemetry.io/otel/trace"
)

func Trace(tracer trace.Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipTracingPaths(r) {
				next.ServeHTTP(w, r)
				return
			}

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// we set span name later after calling the next handler
			// https://github.com/go-chi/chi/blob/master/context.go#L117-L118
			ctx, span := tracer.Start(ctx, "unknown", trace.WithAttributes(
				semconv.HTTPURLKey.String(r.RequestURI),
				semconv.HTTPMethodKey.String(r.Method),
			), trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()

			r = r.WithContext(ctx)
			next.ServeHTTP(ww, r)

			routePattern := chi.RouteContext(ctx).RoutePattern()
			if routePattern == "" {
				routePattern = "<unknown>"
			}

			spanName := fmt.Sprintf("%s %s", r.Method, routePattern)
			span.SetName(spanName)

			status := ww.Status()
			span.SetAttributes(semconv.HTTPStatusCodeKey.Int(status))
			if status >= 400 {
				span.SetStatus(codes.Error, fmt.Sprintf("error with HTTP status code %d", status))
			}
		})
	}
}

var skipPaths = map[string]struct{}{
	"/metrics":          {},
	"/docs":             {},
	"/docs/openapi.yml": {},
}

func skipTracingPaths(r *http.Request) bool {
	_, ok := skipPaths[r.URL.Path]
	return ok
}
