package server

import (
	"github.com/gin-gonic/gin"
	"github.com/playmixer/short-link/internal/logger"
	"github.com/playmixer/short-link/internal/storage"
)

type Logger interface {
	INFO(t ...any)
	WARN(t ...any)
	ERROR(t ...any)
	DEBUG(t ...any)
}

type Storage interface {
	Add(key, value string)
	Get(key string) (string, error)
}

type Server struct {
	addr  string
	port  string
	log   Logger
	store Storage
}

type Option func(s *Server)

func New(options ...Option) *Server {
	srv := &Server{
		addr:  "localhost",
		port:  "8080",
		log:   logger.New(),
		store: storage.New(),
	}

	for _, opt := range options {
		opt(srv)
	}

	return srv
}

func OptionAddr(addr string) func(s *Server) {
	return func(s *Server) {
		s.addr = addr
	}
}

func OptionPort(port string) func(s *Server) {
	return func(s *Server) {
		s.port = port
	}
}

func OptionLogger(logger Logger) func(s *Server) {
	return func(s *Server) {
		s.log = logger
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
	return r.Run()
}
