package rest

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/playmixer/short-link/internal/adapters/models"
	"go.uber.org/zap"
)

const (
	ContentLength   string = "Content-Length"
	ContentType     string = "Content-Type"
	ApplicationJSON string = "application/json"

	CookieNameUserID string = "user_id"
)

type Shortner interface {
	Shorty(ctx context.Context, userID, link string) (string, error)
	ShortyBatch(ctx context.Context, userID string, links []models.ShortenBatchRequest) (
		[]models.ShortenBatchResponse,
		error,
	)
	GetURL(ctx context.Context, short string) (string, error)
	GetAllURL(ctx context.Context, userID string) ([]models.ShortenURL, error)
	PingStore(ctx context.Context) error
}

type Server struct {
	log       *zap.Logger
	addr      string
	short     Shortner
	baseURL   string
	secretKey []byte
}

type Option func(s *Server)

func New(short Shortner, options ...Option) *Server {
	srv := &Server{
		addr:      "localhost:8080",
		short:     short,
		log:       zap.NewNop(),
		secretKey: []byte("rest_secret_key"),
	}

	for _, opt := range options {
		opt(srv)
	}

	return srv
}

func BaseURL(url string) func(*Server) {
	return func(s *Server) {
		s.baseURL = url
	}
}

func Addr(addr string) func(s *Server) {
	return func(s *Server) {
		s.addr = addr
	}
}

func Logger(log *zap.Logger) func(s *Server) {
	return func(s *Server) {
		s.log = log
	}
}

func SecretKey(secret []byte) Option {
	return func(s *Server) {
		s.secretKey = secret
	}
}

func (s *Server) SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(
		s.Logger(),
		s.GzipDecompress(),
	)

	auth := r.Group("/")
	{
		auth.Use(s.CheckCookies())
		auth.POST("/", s.handlerMain)
		auth.GET("/:id", s.handlerShort)
		auth.GET("/ping", s.handlerPing)

		api := auth.Group("/api")
		api.Use(s.GzipCompress())
		{
			api.POST("/shorten", s.handlerAPIShorten)
			api.POST("/shorten/batch", s.handlerAPIShortenBatch)
		}
	}

	userApi := r.Group("/api/user")
	userApi.Use(
		s.GzipCompress(),
		s.Auth(),
	)
	{
		userApi.GET("/urls", s.handlerAPIGetUserURLs)
	}

	return r
}

func (s *Server) Run() error {
	r := s.SetupRouter()
	if err := r.Run(s.addr); err != nil {
		return fmt.Errorf("server has failed: %w", err)
	}
	return nil
}

func (s *Server) baseLink(short string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, short)
}

func (s *Server) SignCookie(data string) string {
	h := hmac.New(sha256.New, s.secretKey)
	h.Write([]byte(data))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s.%s", data, signature)
}

func (s *Server) verifyCookie(signedData string) (string, bool) {
	parts := strings.Split(signedData, ".")
	cookieCountPart := 2
	if len(parts) != cookieCountPart {
		return "", false
	}
	data, signature := parts[0], parts[1]

	expectedSignature := s.SignCookie(data)
	return data, hmac.Equal([]byte(signature), []byte(strings.Split(expectedSignature, ".")[1]))
}
