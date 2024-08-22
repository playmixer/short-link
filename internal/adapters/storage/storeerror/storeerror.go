package storeerror

import "errors"

// Ошибки.
var (
	ErrNotFoundKey       = errors.New("not found value by key")
	ErrNotUnique         = errors.New("row is not unique")
	ErrDuplicateShortURL = errors.New("short url is duplicate")
	ErrShortURLDeleted   = errors.New("short url was deleted")
)
