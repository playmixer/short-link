package shortner

import (
	"net/url"

	"github.com/playmixer/short-link/pkg/util"
)

type ShortI interface {
	Shorty(url string) (string, error)
	GetUrl(short string) (string, error)
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
	_, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	sLink := util.RandomString(6)
	err = s.store.Set(sLink, link)
	if err != nil {
		return "", err
	}
	return sLink, nil
}

func (s *Shortner) GetUrl(short string) (string, error) {

	url, err := s.store.Get(short)
	if err != nil {
		return "", err
	}
	return url, err
}
