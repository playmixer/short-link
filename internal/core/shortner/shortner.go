package shortner

import (
	"fmt"
	"net/url"

	"github.com/playmixer/short-link/pkg/util"
)

var (
	LengthShortLink uint = 6
)

type ShortI interface {
	Shorty(url string) (string, error)
	GetURL(short string) (string, error)
}

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

	sLink := util.RandomString(LengthShortLink)
	err := s.store.Set(sLink, link)
	if err != nil {
		return "", fmt.Errorf("error setting link: %w", err)
	}
	return sLink, nil
}

func (s *Shortner) GetURL(short string) (string, error) {
	link, err := s.store.Get(short)
	if err != nil {
		return "", fmt.Errorf("error getting link: %w", err)
	}
	return link, nil
}
