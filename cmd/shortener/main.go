package main

import (
	"log"

	"github.com/playmixer/short-link/internal/adapters/api/rest"
	"github.com/playmixer/short-link/internal/adapters/config"
	"github.com/playmixer/short-link/internal/adapters/storage"
	"github.com/playmixer/short-link/internal/core/shortner"
)

func main() {

	cfg := config.Init()

	store, err := storage.NewStore(&cfg.Store)
	if err != nil {
		panic(err)
	}
	short := shortner.New(store)

	srv := rest.New(
		short,
		rest.Addr(cfg.Api.Rest.Addr),
		rest.BaseUrl(cfg.BaseUrl),
	)
	log.Fatal(srv.Run())
}
