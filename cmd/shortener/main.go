package main

import (
	"log"

	"github.com/playmixer/short-link/internal/server"
)

func main() {

	srv := server.New(
		server.OptionPort("8080"),
		server.OptionAddr("localhost"),
	)
	log.Fatal(srv.Run())
}
