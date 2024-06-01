package rest

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	ContentLength string = "Content-Length"
	ContentType   string = "Content-Type"
)

type Shortner interface {
	Shorty(ctx context.Context, link string) (string, error)
	GetURL(ctx context.Context, short string) (string, error)
}

type Server struct {
	log     *zap.Logger
	addr    string
	short   Shortner
	baseURL string
}

type Option func(s *Server)

func New(short Shortner, options ...Option) *Server {
	srv := &Server{
		addr:  "localhost:8080",
		short: short,
		log:   zap.NewNop(),
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

func (s *Server) SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(
		s.Logger(),
		s.GzipDecompress(),
	)
	r.POST("/", s.handlerMain)
	r.GET("/:id", s.handlerShort)
	r.GET("/ping", s.handlerPing)

	api := r.Group("/api")
	api.Use(s.GzipCompress())
	{
		api.POST("/shorten", s.handlerAPIShorten)
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
