package rest

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/playmixer/short-link/internal/adapters/models"
	"go.uber.org/zap"
)

const (
	ContentLength   string = "Content-Length"
	ContentType     string = "Content-Type"
	ApplicationJSON string = "application/json"

	CookieNameUserID string = "token"
)

var (
	errInvalidAuthCookie = errors.New("invalid authorization cookie")
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
	DeleteShortURLs(ctx context.Context, shorts []models.ShortLink) error
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

	userAPI := r.Group("/api/user")
	userAPI.Use(
		s.GzipCompress(),
		s.Auth(),
	)
	{
		userAPI.GET("/urls", s.handlerAPIGetUserURLs)
		userAPI.DELETE("/urls", s.handlerAPIDeleteUserURLs)
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
