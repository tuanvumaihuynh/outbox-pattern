package config

import "time"

type Relay struct {
	BatchSize uint32        `env:"RELAY_BATCH_SIZE" envDefault:"100"`
	Interval  time.Duration `env:"RELAY_INTERVAL" envDefault:"1s"`
}
