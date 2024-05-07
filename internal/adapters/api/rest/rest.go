package rest

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type Store interface {
	Set(key, value string)
	Get(key string) (string, error)
}

type Shortner interface {
	Shorty(link string) (string, error)
	GetURL(short string) (string, error)
}

type Server struct {
	addr    string
	short   Shortner
	baseURL string
}

type Option func(s *Server)

func New(short Shortner, options ...Option) *Server {
	srv := &Server{
		addr:  "localhost:8080",
		short: short,
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

func (s *Server) SetupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/", s.mainHandle)
	r.GET("/:id", s.shortHandle)

	return r
}

func (s *Server) Run() error {
	r := s.SetupRouter()
	if err := r.Run(s.addr); err != nil {
		return fmt.Errorf("server has failed: %w", err)
	}
	return nil
}
