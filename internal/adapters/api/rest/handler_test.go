package rest_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/playmixer/short-link/internal/adapters/api/rest"
	"github.com/playmixer/short-link/internal/adapters/auth"
	"github.com/playmixer/short-link/internal/adapters/config"
	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/storage"
	"github.com/playmixer/short-link/internal/adapters/storage/memory"
	"github.com/playmixer/short-link/internal/core/shortner"
)

var (
	cfg *config.Config
)

func initConfig(t *testing.T) {
	t.Helper()
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
		{
			name: "duplicate",
			want: struct {
				StatusCode  int
				Response    string
				Request     string
				ContentType string
			}{
				StatusCode:  http.StatusConflict,
				Response:    "",
				Request:     "https://practicum.yandex.ru/",
				ContentType: "text/plain",
			},
		},
	}

	store, err := storage.NewStore(context.Background(), &storage.Config{Memory: &memory.Config{}}, zap.NewNop())
	if err != nil {
		t.Errorf("failed initialize storage: %v", err)
		return
	}
	authManager, err := auth.New(auth.SetSecretKey([]byte("")))
	if err != nil {
		t.Errorf("failed initializa auth manager: %v", err)
		return
	}
	s := shortner.New(context.Background(), store)
	srv := rest.New(s, authManager, rest.Addr(cfg.API.Rest.Addr), rest.BaseURL(cfg.API.BaseURL))
	router := srv.SetupRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			body := strings.NewReader(tt.want.Request)
			r := httptest.NewRequest(http.MethodPost, "/", body)

			signedCookie, err := authManager.CreateJWT("1")
			require.NoError(t, err)
			r.AddCookie(&http.Cookie{
				Name:  rest.CookieNameUserID,
				Value: signedCookie,
				Path:  "/",
			})

			router.ServeHTTP(w, r)

			result := w.Result()
			assert.Equal(t, tt.want.StatusCode, result.StatusCode)
			assert.Equal(t, tt.want.ContentType, result.Header.Get("Content-type"))
			b, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			if result.StatusCode == http.StatusCreated {
				_, err = url.ParseRequestURI(string(b))
				require.NoError(t, err)
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

	store, err := storage.NewStore(context.Background(), &storage.Config{Memory: &memory.Config{}}, zap.NewNop())
	if err != nil {
		t.Errorf("failed initialize storage: %v", err)
		return
	}
	authManager, err := auth.New(auth.SetSecretKey([]byte("")))
	if err != nil {
		t.Errorf("failed initializa auth manager: %v", err)
		return
	}
	s := shortner.New(context.Background(), store)
	srv := rest.New(s, authManager, rest.Addr(cfg.API.Rest.Addr), rest.BaseURL(cfg.API.BaseURL))
	router := srv.SetupRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/QW23qq", http.NoBody)

			signedCookie, err := authManager.CreateJWT("1")
			require.NoError(t, err)
			r.AddCookie(&http.Cookie{
				Name:  rest.CookieNameUserID,
				Value: signedCookie,
				Path:  "/",
			})

			router.ServeHTTP(w, r)

			result := w.Result()
			assert.Equal(t, tt.want.StatusCode, result.StatusCode)
			assert.Equal(t, tt.want.ContentType, result.Header.Get("Content-type"))
			_ = result.Body.Close()
		})
	}
}

func Test_apiShorten(t *testing.T) {
	initConfig(t)
	fmt.Println("a=", cfg.API.Rest.Addr)
	type tRequest struct {
		URL string `json:"url"`
	}
	tests := []struct {
		name string
		want struct {
			StatusCode  int
			Response    string
			Request     tRequest
			ContentType string
		}
	}{
		{
			name: "empty body",
			want: struct {
				StatusCode  int
				Response    string
				Request     tRequest
				ContentType string
			}{
				StatusCode:  http.StatusBadRequest,
				Response:    "",
				Request:     tRequest{URL: ""},
				ContentType: "",
			},
		},
		{
			name: "no valid link",
			want: struct {
				StatusCode  int
				Response    string
				Request     tRequest
				ContentType string
			}{
				StatusCode:  http.StatusBadRequest,
				Response:    "",
				Request:     tRequest{URL: "test?id=qweq"},
				ContentType: "",
			},
		},
		{
			name: "valid link",
			want: struct {
				StatusCode  int
				Response    string
				Request     tRequest
				ContentType string
			}{
				StatusCode:  http.StatusCreated,
				Response:    "",
				Request:     tRequest{URL: "https://practicum.yandex.ru/"},
				ContentType: "application/json",
			},
		},
		{
			name: "conflict",
			want: struct {
				StatusCode  int
				Response    string
				Request     tRequest
				ContentType string
			}{
				StatusCode:  http.StatusConflict,
				Response:    "",
				Request:     tRequest{URL: "https://practicum.yandex.ru/"},
				ContentType: "application/json",
			},
		},
	}

	store, err := storage.NewStore(context.Background(), &storage.Config{Memory: &memory.Config{}}, zap.NewNop())
	if err != nil {
		t.Errorf("failed initialize storage: %v", err)
		return
	}
	authManager, err := auth.New(auth.SetSecretKey([]byte("")))
	if err != nil {
		t.Errorf("failed initializa auth manager: %v", err)
		return
	}
	s := shortner.New(context.Background(), store)
	srv := rest.New(s, authManager, rest.Addr(cfg.API.Rest.Addr), rest.BaseURL(cfg.API.BaseURL))
	router := srv.SetupRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			reqBody, err := json.Marshal(tt.want.Request)
			if err != nil {
				require.NoError(t, err)
			}
			body := strings.NewReader(string(reqBody))
			r := httptest.NewRequest(http.MethodPost, "/api/shorten", body)

			signedCookie, err := authManager.CreateJWT("1")
			require.NoError(t, err)
			r.AddCookie(&http.Cookie{
				Name:  rest.CookieNameUserID,
				Value: signedCookie,
				Path:  "/",
			})

			router.ServeHTTP(w, r)

			result := w.Result()
			assert.Equal(t, tt.want.StatusCode, result.StatusCode)
			cntnt := result.Header.Get("Content-type")
			assert.Equal(t, tt.want.ContentType, cntnt)
			b, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
			require.Equal(t, result.StatusCode, tt.want.StatusCode)
			if result.StatusCode == http.StatusCreated && err != nil {
				var res struct {
					Result string `json:"result"`
				}
				err = json.Unmarshal(b, &res)
				require.NoError(t, err)
				_, err = url.ParseRequestURI(res.Result)
				require.NoError(t, err)
			}
		})
	}
}

func Test_apiPing(t *testing.T) {
	initConfig(t)
	tests := []struct {
		name string
		want struct {
			StatusCode  int
			Response    string
			ContentType string
		}
	}{
		{
			name: "ok",
			want: struct {
				StatusCode  int
				Response    string
				ContentType string
			}{
				StatusCode:  http.StatusOK,
				Response:    "",
				ContentType: "",
			},
		},
	}

	store, err := storage.NewStore(context.Background(), &storage.Config{Memory: &memory.Config{}}, zap.NewNop())
	if err != nil {
		t.Errorf("failed initialize storage: %v", err)
		return
	}
	authManager, err := auth.New(auth.SetSecretKey([]byte("")))
	if err != nil {
		t.Errorf("failed initializa auth manager: %v", err)
		return
	}
	s := shortner.New(context.Background(), store)
	srv := rest.New(s, authManager, rest.Addr(cfg.API.Rest.Addr), rest.BaseURL(cfg.API.BaseURL))
	router := srv.SetupRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/ping", http.NoBody)

			signedCookie, err := authManager.CreateJWT("1")
			require.NoError(t, err)
			r.AddCookie(&http.Cookie{
				Name:  rest.CookieNameUserID,
				Value: signedCookie,
				Path:  "/",
			})

			router.ServeHTTP(w, r)

			result := w.Result()
			assert.Equal(t, tt.want.StatusCode, result.StatusCode)
			cntnt := result.Header.Get("Content-type")
			assert.Equal(t, tt.want.ContentType, cntnt)
			b, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
			require.Equal(t, result.StatusCode, tt.want.StatusCode)
			if result.StatusCode == http.StatusCreated && err != nil {
				var res struct {
					Result string `json:"result"`
				}
				err = json.Unmarshal(b, &res)
				require.NoError(t, err)
				_, err = url.ParseRequestURI(res.Result)
				require.NoError(t, err)
			}
		})
	}
}

func Test_apiShortenBatch(t *testing.T) {
	initConfig(t)
	type tWant struct {
		StatusCode  int
		Response    []models.ShortenBatchResponse
		ContentType string
	}
	tests := []struct {
		name    string
		request []models.ShortenBatchRequest
		want    tWant
	}{
		{
			name: "default",
			request: []models.ShortenBatchRequest{
				{CorrelationID: "1", OriginalURL: "https://github.com/"},
			},
			want: tWant{
				StatusCode: http.StatusCreated,
				Response: []models.ShortenBatchResponse{
					{CorrelationID: "1"},
				},
				ContentType: rest.ApplicationJSON,
			},
		},
		{
			name: "url invalid",
			request: []models.ShortenBatchRequest{
				{CorrelationID: "1", OriginalURL: "github.com/"},
			},
			want: tWant{
				StatusCode: http.StatusBadRequest,
				Response: []models.ShortenBatchResponse{
					{CorrelationID: "1"},
				},
				ContentType: "",
			},
		},
	}

	store, err := storage.NewStore(context.Background(), &storage.Config{Memory: &memory.Config{}}, zap.NewNop())
	if err != nil {
		t.Errorf("failed initialize storage: %v", err)
		return
	}
	authManager, err := auth.New(auth.SetSecretKey([]byte("")))
	if err != nil {
		t.Errorf("failed initializa auth manager: %v", err)
		return
	}
	s := shortner.New(context.Background(), store)
	srv := rest.New(s, authManager, rest.Addr(cfg.API.Rest.Addr), rest.BaseURL(cfg.API.BaseURL))
	router := srv.SetupRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			reqBody, err := json.Marshal(tt.request)
			if err != nil {
				require.NoError(t, err)
			}
			body := strings.NewReader(string(reqBody))
			r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", body)

			signedCookie, err := authManager.CreateJWT("1")
			require.NoError(t, err)
			r.AddCookie(&http.Cookie{
				Name:  rest.CookieNameUserID,
				Value: signedCookie,
				Path:  "/",
			})

			router.ServeHTTP(w, r)

			result := w.Result()
			assert.Equal(t, tt.want.StatusCode, result.StatusCode)
			cntnt := result.Header.Get("Content-type")
			assert.Equal(t, tt.want.ContentType, cntnt)
			b, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
			require.Equal(t, result.StatusCode, tt.want.StatusCode)
			if result.StatusCode == http.StatusCreated && err != nil {
				var res struct {
					Result string `json:"result"`
				}
				err = json.Unmarshal(b, &res)
				require.NoError(t, err)
				_, err = url.ParseRequestURI(res.Result)
				require.NoError(t, err)
			}
		})
	}
}

func Test_apiDeleteUserURLs(t *testing.T) {
	initConfig(t)
	type tWant struct {
		StatusCode  int
		ContentType string
	}
	tests := []struct {
		name    string
		request []string
		want    tWant
	}{
		{
			name: "default",
			request: []string{
				"QWE1123q",
			},
			want: tWant{
				StatusCode:  http.StatusAccepted,
				ContentType: "",
			},
		},
		{
			name:    "empty",
			request: []string{},
			want: tWant{
				StatusCode:  http.StatusAccepted,
				ContentType: "",
			},
		},
	}

	store, err := storage.NewStore(context.Background(), &storage.Config{Memory: &memory.Config{}}, zap.NewNop())
	if err != nil {
		t.Errorf("failed initialize storage: %v", err)
		return
	}
	authManager, err := auth.New(auth.SetSecretKey([]byte("")))
	if err != nil {
		t.Errorf("failed initializa auth manager: %v", err)
		return
	}
	s := shortner.New(context.Background(), store)
	srv := rest.New(s, authManager, rest.Addr(cfg.API.Rest.Addr), rest.BaseURL(cfg.API.BaseURL))
	router := srv.SetupRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			reqBody, err := json.Marshal(tt.request)
			if err != nil {
				require.NoError(t, err)
			}
			body := strings.NewReader(string(reqBody))
			r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", body)

			if tt.want.StatusCode != http.StatusUnauthorized {
				signedCookie, err := authManager.CreateJWT("1")
				require.NoError(t, err)
				r.AddCookie(&http.Cookie{
					Name:  rest.CookieNameUserID,
					Value: signedCookie,
					Path:  "/",
				})
			}

			router.ServeHTTP(w, r)

			result := w.Result()
			assert.Equal(t, tt.want.StatusCode, result.StatusCode)
			cntnt := result.Header.Get("Content-type")
			assert.Equal(t, tt.want.ContentType, cntnt)
			b, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
			require.Equal(t, result.StatusCode, tt.want.StatusCode)
			if result.StatusCode == http.StatusCreated && err != nil {
				var res struct {
					Result string `json:"result"`
				}
				err = json.Unmarshal(b, &res)
				require.NoError(t, err)
				_, err = url.ParseRequestURI(res.Result)
				require.NoError(t, err)
			}
		})
	}
}

func Test_Gzip(t *testing.T) {
	initConfig(t)
	fmt.Println("a=", cfg.API.Rest.Addr)
	tests := []struct {
		name string
		want struct {
			StatusCode  int
			Request     []byte
			ContentType string
		}
	}{
		{
			name: "zip1",
			want: struct {
				StatusCode  int
				Request     []byte
				ContentType string
			}{
				StatusCode:  http.StatusCreated,
				Request:     []byte(`{"url": "https://practicum.yandex.ru/"}`),
				ContentType: "application/json",
			},
		},
	}

	store, err := storage.NewStore(context.Background(), &storage.Config{Memory: &memory.Config{}}, zap.NewNop())
	if err != nil {
		t.Errorf("failed initialize storage: %v", err)
		return
	}
	authManager, err := auth.New(auth.SetSecretKey([]byte("")))
	if err != nil {
		t.Errorf("failed initializa auth manager: %v", err)
		return
	}
	s := shortner.New(context.Background(), store)
	srv := rest.New(s, authManager, rest.Addr(cfg.API.Rest.Addr), rest.BaseURL(cfg.API.BaseURL))
	router := srv.SetupRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			var buf bytes.Buffer

			gz := gzip.NewWriter(&buf)
			_, err = gz.Write(tt.want.Request)
			if err != nil {
				t.Fatal(err)
			}
			_ = gz.Close()
			r := httptest.NewRequest(http.MethodPost, "/api/shorten", &buf)
			r.Header.Add("Content-Type", "application/json")
			r.Header.Add("Content-Encoding", "gzip")
			r.Header.Add("Accept-Encoding", "gzip")

			signedCookie, err := authManager.CreateJWT("1")
			require.NoError(t, err)
			r.AddCookie(&http.Cookie{
				Name:  rest.CookieNameUserID,
				Value: signedCookie,
				Path:  "/",
			})

			router.ServeHTTP(w, r)

			result := w.Result()
			assert.Equal(t, tt.want.StatusCode, result.StatusCode)
			assert.Equal(t, tt.want.ContentType, result.Header.Get("Content-Type"))

			require.Equal(t, result.StatusCode, tt.want.StatusCode)
			if result.StatusCode == http.StatusCreated && err == nil {
				gr, err := gzip.NewReader(result.Body)
				require.NoError(t, err)
				defer func() { _ = gr.Close() }()
				err = result.Body.Close()
				require.NoError(t, err)
				b, err := io.ReadAll(gr)
				require.NoError(t, err)

				var res struct {
					Result string `json:"result"`
				}

				err = json.Unmarshal(b, &res)
				if err != nil {
					t.Fatal(gr.Extra)
				}
				require.NoError(t, err)
				_, err = url.ParseRequestURI(res.Result)
				require.NoError(t, err)
			}
		})
	}
}

func TestServer_handlerAPIGetUserURLs(t *testing.T) {
	initConfig(t)
	fmt.Println("a=", cfg.API.Rest.Addr)
	tests := []struct {
		name string
		want struct {
			StatusCode  int
			ContentType string
		}
	}{
		{
			name: "getAll",
			want: struct {
				StatusCode  int
				ContentType string
			}{
				StatusCode:  http.StatusNoContent,
				ContentType: "",
			},
		},
	}

	store, err := storage.NewStore(context.Background(), &storage.Config{Memory: &memory.Config{}}, zap.NewNop())
	if err != nil {
		t.Errorf("failed initialize storage: %v", err)
		return
	}
	authManager, err := auth.New(auth.SetSecretKey([]byte("")))
	if err != nil {
		t.Errorf("failed initializa auth manager: %v", err)
		return
	}
	s := shortner.New(context.Background(), store)
	srv := rest.New(s, authManager, rest.Addr(cfg.API.Rest.Addr), rest.BaseURL(cfg.API.BaseURL))
	router := srv.SetupRouter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/user/urls", http.NoBody)
			r.Header.Add("Accept-Encoding", "gzip")

			signedCookie, err := authManager.CreateJWT("1")
			require.NoError(t, err)
			r.AddCookie(&http.Cookie{
				Name:  rest.CookieNameUserID,
				Value: signedCookie,
				Path:  "/",
			})

			router.ServeHTTP(w, r)

			result := w.Result()
			assert.Equal(t, tt.want.StatusCode, result.StatusCode)
			assert.Equal(t, tt.want.ContentType, result.Header.Get("Content-Type"))
			err = result.Body.Close()
			require.NoError(t, err)

			require.Equal(t, result.StatusCode, tt.want.StatusCode)
		})
	}
}
