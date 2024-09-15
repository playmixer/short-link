package rest

// Config конфигурация REST сервиса.
type Config struct {
	Addr        string `env:"SERVER_ADDRESS"`
	SecretKey   string `env:"SECRET_KEY"`
	HTTPSEnable bool   `env:"ENABLE_HTTPS" default:"0"`
}
