package config

type Kafka struct {
	Addresses []string `env:"KAFKA_ADDRESSES,required" envSeparator:","`
	Group     string   `env:"KAFKA_GROUP,required"`
}
