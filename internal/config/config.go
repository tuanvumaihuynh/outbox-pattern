package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// New reads configuration from environment variables (optionally loading a .env file first)
// and unmarshals them into a struct of type T. Returns the populated configuration struct or an error.
func New[T any]() (T, error) {
	var cfg T
	if err := env.Parse(&cfg); err != nil {
		return cfg, fmt.Errorf("parse env: %w", err)
	}

	return cfg, nil
}
