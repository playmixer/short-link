package shortnererror

import "errors"

var (
	// Database storage.
	ErrNotFoundKey       = errors.New("not found value by key")
	ErrNotUnique         = errors.New("row is not unique")
	ErrDuplicateShortURL = errors.New("short url is duplicate")
)
