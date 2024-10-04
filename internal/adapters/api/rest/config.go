package rest

// Config конфигурация REST сервиса.
type Config struct {
	Addr        string `env:"SERVER_ADDRESS"`
	HTTPSEnable bool   `env:"ENABLE_HTTPS"`
}
