package shortner

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/storage/storeerror"
	"github.com/playmixer/short-link/pkg/util"
	"go.uber.org/zap"
)

var (
	LengthShortLink         uint = 6
	NumberOfTryGenShortLink      = 3
	SizeDeleteChanel             = 1024
	HardDeletingDelay            = time.Second * 10
)

type Store interface {
	Get(ctx context.Context, short string) (string, error)
	GetAllURL(ctx context.Context, userID string) ([]models.ShortenURL, error)
	Set(ctx context.Context, userID string, short string, url string) (string, error)
	SetBatch(ctx context.Context, userID string, batch []models.ShortLink) ([]models.ShortLink, error)
	Ping(ctx context.Context) error
	DeleteShortURLs(ctx context.Context, shorts []models.ShortLink) error
	HardDeleteURLs(ctx context.Context) error
}

type Shortner struct {
	store    Store
	deleteCh chan models.ShortLink
	log      *zap.Logger
}

type Option func(*Shortner)

func SetLogger(log *zap.Logger) Option {
	return func(s *Shortner) {
		s.log = log
	}
}

func New(ctx context.Context, s Store, options ...Option) *Shortner {
	sh := &Shortner{
		store:    s,
		deleteCh: make(chan models.ShortLink, SizeDeleteChanel),
		log:      zap.NewNop(),
	}

	for _, opt := range options {
		opt(sh)
	}

	go sh.workerDeleteingShorts(ctx)

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

func (s *Shortner) DeleteShortURLs(ctx context.Context, shorts []models.ShortLink) error {
	err := s.store.DeleteShortURLs(ctx, shorts)
	if err != nil {
		return fmt.Errorf("failed delete short URLs: %w", err)
	}
	return nil
}

func (s *Shortner) workerDeleteingShorts(ctx context.Context) {
	s.log.Debug("start delete short proccessor")
	tick := time.NewTicker(HardDeletingDelay)

	for {
		select {
		case <-ctx.Done():
			s.log.Debug("ended worker `workerDeleteingShorts`")
			return
		case <-tick.C:
			err := s.store.HardDeleteURLs(ctx)
			if err != nil {
				s.log.Error("failed delete short URLs", zap.Error(err))
				continue
			}
		}
	}
}
