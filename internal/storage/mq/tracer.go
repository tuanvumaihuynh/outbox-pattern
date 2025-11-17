package mq

import (
	"github.com/twmb/franz-go/plugin/kotel"
	"go.opentelemetry.io/otel"
)

var (
	tracer  = otel.Tracer("internal/storage/mq")
	kTracer = kotel.NewTracer()
)
