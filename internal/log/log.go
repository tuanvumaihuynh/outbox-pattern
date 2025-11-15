package log

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"

	"github.com/tuanvumaihuynh/outbox-pattern/internal/config"
)

// NewSlogLogger creates a new slog logger with the given configuration.
func NewSlogLogger(cfg config.Log) *slog.Logger {
	var handler slog.Handler

	if cfg.Format == config.LogFormatJSON {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     cfg.Level,
			AddSource: cfg.AddSource,
		})
	} else {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      cfg.Level,
			AddSource:  cfg.AddSource,
			TimeFormat: time.RFC3339,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Value.Kind() == slog.KindAny {
					if _, ok := a.Value.Any().(error); ok {
						return tint.Attr(9, a)
					}
				}
				return a
			},
		})
	}

	handler = newEnrichedHandler(handler)
	log := slog.New(handler)
	slog.SetDefault(log)

	return log
}
