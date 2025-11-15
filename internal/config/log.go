package config

import (
	"fmt"
	"log/slog"
	"strings"
)

type Log struct {
	Format    LogFormat  `env:"LOG_FORMAT" envDefault:"JSON"`
	Level     slog.Level `env:"LOG_LEVEL" envDefault:"INFO"`
	AddSource bool       `env:"LOG_ADD_SOURCE" envDefault:"true"`
}

// LogFormat represents the logging format (JSON or Text).
type LogFormat uint8

// String returns the string representation of the log format.
func (f LogFormat) String() string {
	return []string{"JSON", "TEXT"}[f]
}

const (
	LogFormatJSON LogFormat = iota
	LogFormatText
)

// UnmarshalText implements [encoding.TextUnmarshaler].
// It unmarshals the text to a log format.
func (f *LogFormat) UnmarshalText(text []byte) error {
	switch strings.ToUpper(string(text)) {
	case "JSON":
		*f = LogFormatJSON
	case "TEXT":
		*f = LogFormatText
	default:
		return fmt.Errorf("unknown log format: %s", text)
	}
	return nil
}

func (f LogFormat) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}
