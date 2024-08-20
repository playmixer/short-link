package file

// Config настройка файлового хранилища.
type Config struct {
	StoragePath string `env:"FILE_STORAGE_PATH"`
}
