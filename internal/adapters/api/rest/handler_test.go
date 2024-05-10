package rest_test

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/playmixer/short-link/internal/adapters/api/rest"
	"github.com/playmixer/short-link/internal/adapters/config"
	"github.com/playmixer/short-link/internal/adapters/storage"
	"github.com/playmixer/short-link/internal/adapters/storage/memory"
	"github.com/playmixer/short-link/internal/core/shortner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	cfg *config.Config
)

func initConfig(t *testing.T) {
	if cfg != nil {
		return
	}
	var err error
	cfg, err = config.Init()
	if err != nil {
		t.Fatalf("failed initialize config: %v", err)
	}

	flag.Parse()
}

func Test_mainHandle(t *testing.T) {
	initConfig(t)
	fmt.Println("a=", cfg.API.Rest.Addr)
	tests := []struct {
		name string
		want struct {
			StatusCode  int
			Response    string
			Request     string
			ContentType string
		}
	}{
		{
			name: "empty body",
			want: struct {
				StatusCode  int
				Response    string
				Request     string
				ContentType string
			}{
				StatusCode:  http.StatusBadRequest,
				Response:    "",
				Request:     "",
				ContentType: "text/plain",
			},
		},
		{
			name: "no valid link",
			want: struct {
				StatusCode  int
				Response    string
				Request     string
				ContentType string
			}{
				StatusCode:  http.StatusBadRequest,
				Response:    "",
				Request:     "test?id=qweq",
				ContentType: "text/plain",
			},
		},
		{
			name: "valid link",
			want: struct {
				StatusCode  int
				Response    string
				Request     string
				ContentType string
			}{
				StatusCode:  http.StatusCreated,
				Response:    "",
				Request:     "https://practicum.yandex.ru/",
				ContentType: "text/plain",
			},
		},
	}

	store, err := storage.NewStore(&storage.Config{Memory: &memory.Config{}})
	if err != nil {
		t.Fatalf("failed initialize storage: %v", err)
	}
	s := shortner.New(store)
	srv := rest.New(s, rest.Addr(cfg.API.Rest.Addr), rest.BaseURL(cfg.BaseURL))
	router := srv.SetupRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			body := strings.NewReader(tt.want.Request)
			r := httptest.NewRequest(http.MethodPost, "/", body)
			router.ServeHTTP(w, r)

			result := w.Result()
			assert.Equal(t, tt.want.StatusCode, result.StatusCode)
			assert.Equal(t, tt.want.ContentType, result.Header.Get("Content-type"))
			b, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			if _, err = url.ParseRequestURI(string(b)); result.StatusCode == http.StatusCreated && err != nil {
				t.Fail()
			}
		})
	}
}

func Test_shortHandle(t *testing.T) {
	initConfig(t)
	tests := []struct {
		name string
		want struct {
			StatusCode  int
			Response    string
			Request     string
			ContentType string
		}
	}{
		{
			name: "bad request",
			want: struct {
				StatusCode  int
				Response    string
				Request     string
				ContentType string
			}{
				StatusCode:  http.StatusBadRequest,
				Response:    "",
				Request:     "",
				ContentType: "text/plain",
			},
		},
	}

	store, err := storage.NewStore(&storage.Config{Memory: &memory.Config{}})
	if err != nil {
		t.Fatalf("failed initialize storage: %v", err)
	}
	s := shortner.New(store)
	srv := rest.New(s, rest.Addr(cfg.API.Rest.Addr), rest.BaseURL(cfg.BaseURL))
	router := srv.SetupRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/QW23qq", http.NoBody)
			router.ServeHTTP(w, r)

			result := w.Result()
			assert.Equal(t, tt.want.StatusCode, result.StatusCode)
			assert.Equal(t, tt.want.ContentType, result.Header.Get("Content-type"))
			_ = result.Body.Close()
		})
	}
}
