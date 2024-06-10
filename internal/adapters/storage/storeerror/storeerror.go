package storeerror

import "errors"

var (
	ErrNotFoundKey       = errors.New("not found value by key")
	ErrNotUnique         = errors.New("row is not unique")
	ErrDuplicateShortURL = errors.New("short url is duplicate")
)
