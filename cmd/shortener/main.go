package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"github.com/playmixer/short-link/internal/storage"
)

type Storage interface {
	Add(key, value string)
	Get(key string) (string, error)
}

var (
	store Storage
)

func init() {
	store = storage.New()
}

func mainHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		shortHandle(w, r)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Add("Content-Type", "text/plain")

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = url.ParseRequestURI(string(b))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sLink := randomString()
	store.Add(sLink, string(b))

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("http://localhost:8080/%s", sLink)))
}

func shortHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Add("Content-Type", "text/plain")

	id := r.URL.Path
	id = strings.ReplaceAll(id, "/", "")

	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("page `%s` not found", id)
		return
	}
	url, err := store.Get(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("page not found by id `%s`", id)
		return
	}
	w.Header().Add("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func randomString() string {
	return fmt.Sprintf("%x", rand.Uint32())
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/{id}", shortHandle)
	mux.HandleFunc("/", mainHandle)
	http.ListenAndServe(":8080", mux)
}
