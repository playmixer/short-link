package shortner

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/playmixer/short-link/pkg/util"
)

var (
	LengthShortLink         uint = 6
	NumberOfTryGenShortLink      = 3
)

type Store interface {
	Get(short string) (string, error)
	Set(short string, url string) error
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

func (s *Shortner) Shorty(link string) (string, error) {
	if _, err := url.Parse(link); err != nil {
		return "", fmt.Errorf("error parsing link: %w", err)
	}

	var i int
	for {
		sLink := util.RandomString(LengthShortLink)
		if err := s.store.Set(sLink, link); err == nil {
			return sLink, nil
		}
		i++
		if i >= NumberOfTryGenShortLink {
			break
		}
	}

	return "", errors.New("failed to generate a unique short link")
}

func (s *Shortner) GetURL(short string) (string, error) {
	link, err := s.store.Get(short)
	if err != nil {
		return "", fmt.Errorf("error getting link: %w", err)
	}
	return link, nil
}
