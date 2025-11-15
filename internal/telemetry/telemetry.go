package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/config"
)

type CleanupFunc func(context.Context) error

// InitTracer initializes the OpenTelemetry tracer.
// Should be called at the start of the application to get the tracer set globally.
func InitTracer(ctx context.Context, cfg config.Otel) (CleanupFunc, error) {
	if cfg.CollectorURL == "" {
		// no-op
		return func(context.Context) error {
			return nil
		}, nil
	}

	var secureOpt otlptracegrpc.Option

	if !cfg.Insecure {
		secureOpt = otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	} else {
		secureOpt = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			secureOpt,
			otlptracegrpc.WithEndpoint(cfg.CollectorURL),
			otlptracegrpc.WithHeaders(map[string]string{
				"Authorization": cfg.CollectorAuth,
			}),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create exporter: %w", err)
	}

	resourceAttrs := []attribute.KeyValue{
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("library.language", "go"),
	}

	if cfg.K8sPodName != "" && cfg.K8sNamespace != "" {
		resourceAttrs = append(resourceAttrs, attribute.String("k8s.pod.name", cfg.K8sPodName))
		resourceAttrs = append(resourceAttrs, attribute.String("k8s.namespace", cfg.K8sNamespace))
	}

	resources, err := resource.New(
		ctx,
		resource.WithAttributes(resourceAttrs...),
	)
	if err != nil {
		return nil, fmt.Errorf("set resources: %w", err)
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.TraceIDRatio)),
			sdktrace.WithBatcher(
				exporter,
				sdktrace.WithMaxQueueSize(sdktrace.DefaultMaxQueueSize*10),
				sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize*10),
			),
			sdktrace.WithResource(resources),
		),
	)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	cleanup := func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := exporter.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown OpenTelemetry exporter: %w", err)
		}

		return nil
	}

	return cleanup, nil
}
