package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/shortnererror"
	"go.uber.org/zap"
)

type Store struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

func initTables(ctx context.Context, db *pgxpool.Pool) error {
	exec := `CREATE TABLE IF NOT EXISTS public.short_link (
		id int GENERATED ALWAYS AS IDENTITY NOT NULL,
		original_url varchar NOT NULL,
		short_url varchar NOT NULL,
		CONSTRAINT short_url_pk PRIMARY KEY (short_url)
	);
	CREATE UNIQUE INDEX IF NOT EXISTS short_link_original_url_idx ON public.short_link USING btree (original_url);
	CREATE UNIQUE INDEX IF NOT EXISTS short_link_short_url_idx ON public.short_link USING btree (short_url);`

	_, err := db.Exec(ctx, exec)
	if err != nil {
		return fmt.Errorf("error initialize tables: %w", err)
	}

	return nil
}

func New(ctx context.Context, cfg *Config) (*Store, error) {
	var err error
	s := &Store{
		log: cfg.log,
	}
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed parse config: %w", err)
	}
	s.pool, err = pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed open database: %w", err)
	}
	err = initTables(ctx, s.pool)
	if err != nil {
		return nil, fmt.Errorf("failed initialize tables: %w", err)
	}

	return s, nil
}

func (s *Store) Set(ctx context.Context, short, original string) (output string, err error) {
	_, err = s.pool.Exec(ctx, "insert into short_link (short_url, original_url) values ($1, $2)", short, original)
	var sqlError *pgconn.PgError
	if err != nil && errors.As(err, &sqlError) && pgerrcode.UniqueViolation == sqlError.Code {
		output, err = s.GetByOriginal(ctx, original)
		if err != nil {
			return output, fmt.Errorf("failed select URL %s %w %w", original, err, shortnererror.ErrDuplicateShortURL)
		}
		return output, fmt.Errorf("pgerror: %w: %w", shortnererror.ErrNotUnique, err)
	}
	if err != nil {
		return output, fmt.Errorf("failed setting short url: %w", err)
	}
	return short, nil
}

func (s *Store) Get(ctx context.Context, key string) (string, error) {
	row := s.pool.QueryRow(ctx, "select original_url from short_link where short_url =$1", key)
	var value string
	err := row.Scan(&value)
	if err != nil {
		return "", fmt.Errorf("failed scan url %w", err)
	}
	return value, nil
}

func (s *Store) SetBatch(ctx context.Context, data []models.ShortLink) (output []models.ShortLink, reserr error) {
	output = make([]models.ShortLink, 0)
	sqlString := "insert into short_link (short_url, original_url) values (@short, @original)"
	batch := &pgx.Batch{}

	for _, v := range data {
		args := pgx.NamedArgs{
			"short":    v.ShortURL,
			"original": v.OriginalURL,
		}
		batch.Queue(sqlString, args)
	}

	result := s.pool.SendBatch(ctx, batch)
	defer func() {
		err := result.Close()
		if err != nil {
			s.log.Debug("error closing result batch", zap.Error(err))
		}
	}()

	for _, v := range data {
		_, err := result.Exec()
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				reserr = errors.Join(err, shortnererror.ErrNotUnique)
				v.ShortURL, err = s.GetByOriginal(ctx, v.OriginalURL)
				if err != nil {
					return output, fmt.Errorf("failed select URL %s %w %w", v.OriginalURL, err, shortnererror.ErrDuplicateShortURL)
				}
				return []models.ShortLink{v}, fmt.Errorf("URL %s is not unique: %w", v.OriginalURL, reserr)
			} else {
				return []models.ShortLink{}, fmt.Errorf("failed insert URL %s: %w", v.OriginalURL, err)
			}
		}
		output = append(output, v)
	}

	return output, nil
}

func (s *Store) GetByOriginal(ctx context.Context, original string) (string, error) {
	row := s.pool.QueryRow(ctx, "select short_url from short_link where original_url =$1", original)
	var value string
	err := row.Scan(&value)
	if err != nil {
		return "", fmt.Errorf("failed scan url %w", err)
	}
	return value, nil
}

func (s *Store) Ping(ctx context.Context) error {
	err := s.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed ping database: %w", err)
	}

	return nil
}
