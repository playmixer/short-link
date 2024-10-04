package api

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/playmixer/short-link/internal/adapters/api/grpch"
	"github.com/playmixer/short-link/internal/adapters/api/rest"
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

type API interface {
	Run() error
	Stop()
}

type AuthManager interface {
	VerifyJWT(signedData string) (string, bool)
	CreateJWT(uniqueID string) (string, error)
}

func New(short Shortner, auth AuthManager, lgr *zap.Logger, cfg Config) (API, error) {
	if cfg.GRPC != nil && cfg.GRPC.Addr != "" {
		srv, err := grpch.New(
			short,
			auth,
			grpch.Address(cfg.GRPC.Addr),
			grpch.Logger(lgr),
			grpch.SecretKey([]byte(cfg.SecretKey)),
			grpch.TrustedSubnet(cfg.TrustedSubnet),
		)
		if err != nil {
			return nil, fmt.Errorf("failed initialize grpc server: %w", err)
		}
		return srv, nil
	}
	if cfg.Rest != nil {
		srv := rest.New(
			short,
			auth,
			rest.Addr(cfg.Rest.Addr),
			rest.BaseURL(cfg.BaseURL),
			rest.Logger(lgr),
			rest.SecretKey([]byte(cfg.SecretKey)),
			rest.HTTPSEnable(cfg.Rest.HTTPSEnable),
			rest.TrastedSubnet(cfg.TrustedSubnet),
		)
		return srv, nil
	}
	return nil, errors.New("not found any API")
}
