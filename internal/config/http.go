package config

type HTTP struct {
	Port    uint32 `env:"HTTP_PORT" envDefault:"8000"`
	Swagger bool   `env:"HTTP_SWAGGER" envDefault:"true"`
}
