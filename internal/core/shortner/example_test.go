package shortner_test

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/storage"
	"github.com/playmixer/short-link/internal/adapters/storage/memory"
	"github.com/playmixer/short-link/internal/core/shortner"
)

func ExampleShortner_Shorty() {
	ctx := context.Background()

	store, _ := storage.NewStore(ctx, &storage.Config{Memory: &memory.Config{}}, zap.NewNop())
	_, _ = store.Set(ctx, "1", "VLIWXD", "https://practicum.yandex.ru/")

	s := shortner.New(ctx, store)

	// Сокращаем ссылку.
	shortLink, _ := s.Shorty(ctx, "1", "https://practicum.yandex.ru/")

	fmt.Println(shortLink)
	// Output:
	// VLIWXD
}

func ExampleShortner_GetURL() {
	ctx := context.Background()

	store, _ := storage.NewStore(ctx, &storage.Config{Memory: &memory.Config{}}, zap.NewNop())
	_, _ = store.Set(ctx, "1", "VLIWXD", "https://practicum.yandex.ru/")

	s := shortner.New(ctx, store)

	// Получаем полную ссылку.
	link, _ := s.GetURL(ctx, "VLIWXD")
	fmt.Println(link)
	// Output:
	// https://practicum.yandex.ru/
}

func ExampleShortner_ShortyBatch() {
	ctx := context.Background()

	store, _ := storage.NewStore(ctx, &storage.Config{Memory: &memory.Config{}}, zap.NewNop())
	_, _ = store.Set(ctx, "1", "VLIWXD", "https://practicum.yandex.ru/")

	s := shortner.New(ctx, store)

	// Сохраняем массив ссылок.
	output, _ := s.ShortyBatch(ctx, "1", []models.ShortenBatchRequest{
		{
			CorrelationID: "1",
			OriginalURL:   "https://practicum.yandex.ru/",
		},
	})
	fmt.Println(output)

	// Output:
	// [{1 VLIWXD}]
}

func ExampleShortner_GetAllURL() {
	ctx := context.Background()

	store, _ := storage.NewStore(ctx, &storage.Config{Memory: &memory.Config{}}, zap.NewNop())
	_, _ = store.Set(ctx, "1", "VLIWXD", "https://practicum.yandex.ru/")

	s := shortner.New(ctx, store)

	// Получаем все ссылки пользователя.
	output, _ := s.GetAllURL(ctx, "1")
	fmt.Println(output)

	// Output:
	// [{VLIWXD https://practicum.yandex.ru/}]
}

func ExampleShortner_DeleteShortURLs() {
	ctx := context.Background()

	store, _ := storage.NewStore(ctx, &storage.Config{Memory: &memory.Config{}}, zap.NewNop())
	_, _ = store.Set(ctx, "1", "VLIWXD", "https://practicum.yandex.ru/")

	s := shortner.New(ctx, store)

	// Удаляем сохраненные ссылки.
	err := s.DeleteShortURLs(ctx, []models.ShortLink{
		{
			ShortURL:    "VLIWXD",
			OriginalURL: "https://practicum.yandex.ru/",
			UserID:      "1",
		},
	})
	fmt.Println(err)

	// Output:
	// <nil>
}
