package grpch

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/playmixer/short-link/internal/adapters/api/grpch/proto"
	"github.com/playmixer/short-link/internal/adapters/models"
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

// Server - GRPC API сервер.
type Server struct {
	pb.UnimplementedShortenServer

	log           *zap.Logger
	s             *grpc.Server
	short         Shortner
	auth          AuthManager
	addr          string
	trustedSubnet string
	secretKey     []byte
}

// Option - опции сервера.
type Option func(s *Server)

// Address адрес сервера.
func Address(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

// SecretKey - задает секретный ключ.
func SecretKey(secret []byte) Option {
	return func(s *Server) {
		s.secretKey = secret
	}
}

// Logger - Устанавливает логер.
func Logger(lgr *zap.Logger) Option {
	return func(s *Server) {
		s.log = lgr
	}
}

// TrustedSubnet - подсеть разрешенных адресов.
func TrustedSubnet(subnet string) Option {
	return func(s *Server) {
		s.trustedSubnet = subnet
	}
}

// New создает Server.
func New(short Shortner, auth AuthManager, options ...Option) (*Server, error) {
	srv := &Server{
		short:     short,
		auth:      auth,
		log:       zap.NewNop(),
		secretKey: []byte("rest_secret_key"),
	}
	srv.s = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			srv.interceptorLogger,
		),
	)
	srv.addr = "localhost:8081"

	for _, opt := range options {
		opt(srv)
	}

	return srv, nil
}

// Run запуске сервера.
func (s *Server) Run() error {
	listen, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed listen: %w", err)
	}
	pb.RegisterShortenServer(s.s, s)
	s.log.Info("GRPC service starting...", zap.String("address", s.addr))
	if err := s.s.Serve(listen); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

// Stop остановка сервера.
func (s *Server) Stop() {
	s.s.Stop()
}

func (s *Server) getAuth(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("metadata is not provided")
	}

	values := md["token"]
	if len(values) == 0 {
		return "", errors.New("authorization token is not provided")
	}

	token := values[0]
	var userID string
	if userID, ok = s.auth.VerifyJWT(token); !ok || userID == "" {
		return "", errors.New("authorization token is not valid")
	}

	return userID, nil
}

func (s *Server) getMetadata(ctx context.Context, name string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("metadata is not provided")
	}

	name = strings.ToLower(name)
	values := md[name]
	if len(values) == 0 {
		return "", fmt.Errorf("%s is not provided", name)
	}

	res := values[0]

	return res, nil
}
