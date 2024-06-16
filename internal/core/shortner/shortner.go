package shortner

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/storage/storeerror"
	"github.com/playmixer/short-link/pkg/util"
)

var (
	LengthShortLink         uint = 6
	NumberOfTryGenShortLink      = 3
)

type Store interface {
	Get(ctx context.Context, short string) (string, error)
	GetAllURL(ctx context.Context, userID string) ([]models.ShortenURL, error)
	Set(ctx context.Context, userID string, short string, url string) (string, error)
	SetBatch(ctx context.Context, userID string, batch []models.ShortLink) ([]models.ShortLink, error)
	Ping(ctx context.Context) error
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

func (s *Shortner) Shorty(ctx context.Context, userID, link string) (sLink string, err error) {
	if _, err = url.Parse(link); err != nil {
		return "", fmt.Errorf("error parsing link: %w", err)
	}

	var i int
	for {
		sLink = util.RandomString(LengthShortLink)
		sLink, err = s.store.Set(ctx, userID, sLink, link)
		if err != nil && !errors.Is(err, storeerror.ErrDuplicateShortURL) {
			return sLink, fmt.Errorf("failed setting URL %s: %w", link, err)
		}
		if err == nil {
			return sLink, nil
		}
		i++
		if i >= NumberOfTryGenShortLink {
			break
		}
	}

	return sLink, fmt.Errorf("failed to generate a unique short link: %w", err)
}

func (s *Shortner) GetURL(ctx context.Context, short string) (string, error) {
	link, err := s.store.Get(ctx, short)
	if err != nil {
		return "", fmt.Errorf("error getting link: %w", err)
	}
	return link, nil
}

func (s *Shortner) ShortyBatch(ctx context.Context, userID string, batch []models.ShortenBatchRequest) (
	output []models.ShortenBatchResponse,
	err error,
) {
	payload := make([]models.ShortLink, 0, len(batch))
	for _, batchRequest := range batch {
		short := util.RandomString(LengthShortLink)
		payload = append(payload, models.ShortLink{
			ShortURL:    short,
			OriginalURL: batchRequest.OriginalURL,
		})
	}
	results, err := s.store.SetBatch(ctx, userID, payload)
	output = make([]models.ShortenBatchResponse, 0)

	for i := range results {
		for l := range batch {
			if results[i].OriginalURL == batch[l].OriginalURL {
				output = append(output, models.ShortenBatchResponse{
					CorrelationID: batch[l].CorrelationID,
					ShortURL:      results[i].ShortURL,
				})
				break
			}
		}
	}
	if err != nil {
		return output, fmt.Errorf("failed insert list URLs: %w", err)
	}

	return output, nil
}

func (s *Shortner) PingStore(ctx context.Context) error {
	err := s.store.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed ping storage: %w", err)
	}
	return nil
}

func (s *Shortner) GetAllURL(ctx context.Context, userID string) ([]models.ShortenURL, error) {
	data, err := s.store.GetAllURL(ctx, userID)
	if err != nil {
		return data, fmt.Errorf("failed get all URLs: %w", err)
	}
	return data, nil
}
