package models

// ShortLink модель хранения коротких ссылок.
type ShortLink struct {
	ShortURL    string
	OriginalURL string
	UserID      string
	ID          int64
}
