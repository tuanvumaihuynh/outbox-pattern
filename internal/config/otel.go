package config

type Otel struct {
	ServiceName   string  `env:"OTEL_SERVICE_NAME"`
	CollectorURL  string  `env:"OTEL_COLLECTOR_URL"`
	Insecure      bool    `env:"OTEL_INSECURE"`
	TraceIDRatio  float64 `env:"OTEL_TRACE_ID_RATIO" envDefault:"0.1"`
	CollectorAuth string  `env:"OTEL_COLLECTOR_AUTH"`

	K8sPodName   string `env:"K8S_POD_NAME"`
	K8sNamespace string `env:"K8S_NAMESPACE"`
}
