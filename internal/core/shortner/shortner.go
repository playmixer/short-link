package shortner

import (
	"context"
	"fmt"
	"net/url"

	"github.com/playmixer/short-link/pkg/util"
)

var (
	LengthShortLink         uint = 6
	NumberOfTryGenShortLink      = 3
)

type Store interface {
	Get(ctx context.Context, short string) (string, error)
	Set(ctx context.Context, short string, url string) error
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
