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
	"github.com/golang-jwt/jwt/v5"
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
}

// Server - REST API сервер.
type Server struct {
	log       *zap.Logger
	short     Shortner
	baseURL   string
	secretKey []byte
	s         http.Server
	tlsEnable bool
}

// Option - опции сервера.
type Option func(s *Server)

// New создает Server.
func New(short Shortner, options ...Option) *Server {
	srv := &Server{
		short:     short,
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

// CreateJWT - Создает JWT ключ и записывает в него ID пользователя.
func (s *Server) CreateJWT(uniqueID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uniqueID": uniqueID,
	})
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed signe token: %w", err)
	}

	return tokenString, nil
}

func (s *Server) verifyJWT(signedData string) (string, bool) {
	token, err := jwt.Parse(signedData, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unknown signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		s.log.Debug("failed parse jwt token", zap.Error(err))
		return "", false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if uniqueID, ok := claims["uniqueID"].(string); ok {
			if uniqueID != "" {
				return uniqueID, true
			}
		}
	}

	return "", false
}

func (s *Server) checkAuth(c *gin.Context) (userID string, err error) {
	var ok bool
	cookieUserID, err := c.Request.Cookie(CookieNameUserID)
	if err == nil {
		userID, ok = s.verifyJWT(cookieUserID.Value)
	}
	if err != nil {
		return "", fmt.Errorf("failed reade user cookie: %w %w", errInvalidAuthCookie, err)
	}
	if !ok {
		return "", fmt.Errorf("unverify usercookie: %w", errInvalidAuthCookie)
	}

	return userID, nil
}
