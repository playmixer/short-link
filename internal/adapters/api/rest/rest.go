// Модуль rest предоставляет http сервер и методы взаимодействия с REST API.
package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/pprof"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/playmixer/short-link/internal/adapters/models"
)

// Константы сервиса.
const (
	ContentLength   string = "Content-Length"   // заголовок длины конетента
	ContentType     string = "Content-Type"     // заколовок типа контент
	ApplicationJSON string = "application/json" // json контент

	CookieNameUserID string = "token" // поле хранения токента
)

var (
	errInvalidAuthCookie = errors.New("invalid authorization cookie")

	shutdownDelay = time.Second * 5
)

// Shortner интерфейс взаимодействия с сервисом сокращения ссылок.
type Shortner interface {
	Shorty(ctx context.Context, userID, link string) (string, error)
	ShortyBatch(ctx context.Context, userID string, links []models.ShortenBatchRequest) (
		[]models.ShortenBatchResponse,
		error,
	)
	GetURL(ctx context.Context, short string) (string, error)
	GetAllURL(ctx context.Context, userID string) ([]models.ShortenURL, error)
	PingStore(ctx context.Context) error
	DeleteShortURLs(ctx context.Context, shorts []models.ShortLink) error
	GetState(ctx context.Context) (models.ShortenStats, error)
}

type AuthManager interface {
	VerifyJWT(signedData string) (string, bool)
	CreateJWT(uniqueID string) (string, error)
}

// Server - REST API сервер.
type Server struct {
	log           *zap.Logger
	auth          AuthManager
	short         Shortner
	baseURL       string
	trustedSubnet string
	secretKey     []byte
	s             http.Server
	tlsEnable     bool
}

// Option - опции сервера.
type Option func(s *Server)

// New создает Server.
func New(short Shortner, auth AuthManager, options ...Option) *Server {
	srv := &Server{
		short:     short,
		auth:      auth,
		log:       zap.NewNop(),
		secretKey: []byte("rest_secret_key"),
	}
	srv.s.Addr = "localhost:8080"

	for _, opt := range options {
		opt(srv)
	}

	return srv
}

// BaseURL - Настройка сервера, задает полный путь для сокращенной ссылки.
func BaseURL(url string) func(*Server) {
	return func(s *Server) {
		s.baseURL = url
	}
}

// Addr - Насткройка сервера, задает адрес сервера.
func Addr(addr string) func(s *Server) {
	return func(s *Server) {
		s.s.Addr = addr
	}
}

// Logger - Устанавливает логер.
func Logger(log *zap.Logger) func(s *Server) {
	return func(s *Server) {
		s.log = log
	}
}

// SecretKey - задает секретный ключ.
func SecretKey(secret []byte) Option {
	return func(s *Server) {
		s.secretKey = secret
	}
}

// HTTPSEnable - включает https.
func HTTPSEnable(enable bool) Option {
	return func(s *Server) {
		s.tlsEnable = enable
	}
}

// TrastedSubnet - установка доступной сети.
func TrastedSubnet(subnet string) Option {
	return func(s *Server) {
		s.trustedSubnet = subnet
	}
}

// SetupRouter - создает маршруты.
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

	userAPI := r.Group("/api/user")
	userAPI.Use(
		s.GzipCompress(),
		s.Auth(),
	)
	{
		userAPI.GET("/urls", s.handlerAPIGetUserURLs)
		userAPI.DELETE("/urls", s.handlerAPIDeleteUserURLs)
	}

	interAPI := r.Group("/api/internal")
	interAPI.Use(
		s.TrustedSubnet(),
	)
	{
		interAPI.GET("/stats", s.handlerAPIInternalStats)
	}

	pprof.Register(r, "debug/pprof")

	return r
}

// Run - запускает сервер.
func (s *Server) Run() error {
	s.s.Handler = s.SetupRouter().Handler()
	switch s.tlsEnable {
	case false:
		if err := s.s.ListenAndServe(); err != nil {
			return fmt.Errorf("server has failed: %w", err)
		}
	case true:
		if err := s.s.ListenAndServeTLS("./cert/shortner.crt", "./cert/shortner.key"); err != nil {
			return fmt.Errorf("server has failed: %w", err)
		}
	}
	return nil
}

// Stop - остановка сервера.
func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownDelay)
	defer cancel()
	err := s.s.Shutdown(ctx)
	if err != nil {
		s.log.Error("failed shutdown server", zap.Error(err))
	}
	s.log.Info("Server exiting")
}

func (s *Server) baseLink(short string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, short)
}

func (s *Server) checkAuth(c *gin.Context) (userID string, err error) {
	var ok bool
	cookieUserID, err := c.Request.Cookie(CookieNameUserID)
	if err == nil {
		userID, ok = s.auth.VerifyJWT(cookieUserID.Value)
	}
	if err != nil {
		return "", fmt.Errorf("failed reade user cookie: %w %w", errInvalidAuthCookie, err)
	}
	if !ok {
		return "", fmt.Errorf("unverify usercookie: %w", errInvalidAuthCookie)
	}

	return userID, nil
}
