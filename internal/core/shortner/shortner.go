package shortner

import (
	"context"
	"fmt"
	"net/url"

	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/pkg/util"
)

var (
	LengthShortLink         uint = 6
	NumberOfTryGenShortLink      = 3
)

type Store interface {
	Get(ctx context.Context, short string) (string, error)
	Set(ctx context.Context, short string, url string) error
	SetBatch(ctx context.Context, batch []models.ShortLink) error
}

type Shortner struct {
	store Store
}

type Option func(*Shortner)

func New(s Store, options ...Option) *Shortner {
	sh := &Shortner{
		store: s,
	}

	for _, opt := range options {
		opt(sh)
	}

	return sh
}

func (s *Shortner) Shorty(ctx context.Context, link string) (string, error) {
	var err error
	if _, err = url.Parse(link); err != nil {
		return "", fmt.Errorf("error parsing link: %w", err)
	}

	var i int
	for {
		sLink := util.RandomString(LengthShortLink)
		if err = s.store.Set(ctx, sLink, link); err == nil {
			return sLink, nil
		}
		i++
		if i >= NumberOfTryGenShortLink {
			break
		}
	}

	return "", fmt.Errorf("failed to generate a unique short link: %w", err)
}

func (s *Shortner) GetURL(ctx context.Context, short string) (string, error) {
	link, err := s.store.Get(ctx, short)
	if err != nil {
		return "", fmt.Errorf("error getting link: %w", err)
	}
	return link, nil
}

func (s *Shortner) ShortyBatch(ctx context.Context, batch []models.ShortenBatchRequest) (
	links []models.ShortenBatchResponse,
	err error,
) {
	links = make([]models.ShortenBatchResponse, 0, len(batch))
	payload := make([]models.ShortLink, 0, len(batch))
	for _, batchRequest := range batch {
		short := util.RandomString(LengthShortLink)
		links = append(links, models.ShortenBatchResponse{
			CorrelationID: batchRequest.CorrelationID,
			ShortURL:      short,
		})
		payload = append(payload, models.ShortLink{
			ShortURL:    short,
			OriginalURL: batchRequest.OriginalURL,
		})
	}
	err = s.store.SetBatch(ctx, payload)
	if err != nil {
		return links, fmt.Errorf("failed insert list URLs: %w", err)
	}
	return links, nil
}
