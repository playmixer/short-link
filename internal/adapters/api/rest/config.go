package rest

type Config struct {
	Addr      string `env:"SERVER_ADDRESS"`
	SecretKey string `env:"SECRET_KEY"`
}
