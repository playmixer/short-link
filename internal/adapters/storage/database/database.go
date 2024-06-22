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
	"github.com/playmixer/short-link/internal/adapters/storage/storeerror"
	"go.uber.org/zap"
)

type Store struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

func initTables(ctx context.Context, db *pgxpool.Pool) error {
	exec := `CREATE TABLE IF NOT EXISTS public.short_link (
	id int4 GENERATED ALWAYS AS IDENTITY NOT NULL,
	original_url varchar NOT NULL,
	short_url varchar NOT NULL,
	user_id varchar NOT NULL,
	is_deleted bool DEFAULT false NOT NULL,
	CONSTRAINT short_url_pk PRIMARY KEY (short_url)
);
CREATE UNIQUE INDEX IF NOT EXISTS short_link_original_url_idx ON public.short_link (original_url,user_id);
CREATE UNIQUE INDEX IF NOT EXISTS short_link_short_url_idx ON public.short_link (short_url,user_id);`

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

func (s *Store) Set(ctx context.Context, userID, short, original string) (output string, err error) {
	_, err = s.pool.Exec(
		ctx,
		"insert into short_link (short_url, original_url, user_id) values ($1, $2, $3)",
		short, original, userID,
	)
	var sqlError *pgconn.PgError
	if err != nil {
		if errors.As(err, &sqlError) && pgerrcode.UniqueViolation == sqlError.Code {
			output, err = s.getByOriginal(ctx, userID, original)
			if err != nil {
				return output, fmt.Errorf("failed select URL %s %w %w", original, err, storeerror.ErrDuplicateShortURL)
			}
			return output, fmt.Errorf("pgerror: %w: %w", storeerror.ErrNotUnique, err)
		}
		return output, fmt.Errorf("failed setting short url: %w", err)
	}
	return short, nil
}

func (s *Store) Get(ctx context.Context, short string) (string, error) {
	row := s.pool.QueryRow(ctx,
		"select original_url, is_deleted from short_link where short_url = $1",
		short,
	)
	var value string
	var isDeleted bool
	err := row.Scan(&value, &isDeleted)
	if err != nil {
		return "", fmt.Errorf("failed scan url %w", err)
	}
	if isDeleted {
		return short, storeerror.ErrShortURLDeleted
	}
	return value, nil
}

func (s *Store) SetBatch(ctx context.Context, userID string, data []models.ShortLink) (
	output []models.ShortLink,
	reserr error,
) {
	output = make([]models.ShortLink, 0)
	sqlString := "insert into short_link (short_url, original_url, user_id) values (@short, @original, @user_id)"
	batch := &pgx.Batch{}

	for _, v := range data {
		args := pgx.NamedArgs{
			"short":    v.ShortURL,
			"original": v.OriginalURL,
			"user_id":  userID,
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
				reserr = errors.Join(err, storeerror.ErrNotUnique)
				v.ShortURL, err = s.getByOriginal(ctx, userID, v.OriginalURL)
				if err != nil {
					return output, fmt.Errorf("failed select URL %s %w %w",
						v.OriginalURL, err, storeerror.ErrDuplicateShortURL)
				}
				return []models.ShortLink{v}, fmt.Errorf("URL %s is not unique: %w", v.OriginalURL, reserr)
			}
			return []models.ShortLink{}, fmt.Errorf("failed insert URL %s: %w", v.OriginalURL, err)
		}
		output = append(output, v)
	}

	return output, nil
}

func (s *Store) getByOriginal(ctx context.Context, userID, original string) (string, error) {
	row := s.pool.QueryRow(ctx,
		"select short_url from short_link where original_url =$1 and user_id = $2",
		original, userID,
	)
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

func (s *Store) GetAllURL(ctx context.Context, userID string) ([]models.ShortenURL, error) {
	result := []models.ShortenURL{}
	rows, err := s.pool.Query(ctx, "select short_url, original_url from short_link where user_id = $1", userID)
	if err != nil {
		return result, fmt.Errorf("failed selecting all URLs by user: %w", err)
	}
	for rows.Next() {
		value := models.ShortenURL{}
		err := rows.Scan(&value.ShortURL, &value.OriginalURL)
		if err != nil {
			return result, fmt.Errorf("failed scan url %w", err)
		}
		result = append(result, value)
	}
	return result, nil
}

func (s *Store) DeleteShortURLs(ctx context.Context, shorts []models.ShortLink) error {
	sqlString := `update short_link set is_deleted = true 
where user_id = @user_id and short_url = @short_url and is_deleted = false`
	batch := &pgx.Batch{}

	for _, v := range shorts {
		args := pgx.NamedArgs{
			"short_url": v.ShortURL,
			"user_id":   v.UserID,
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

	for _, v := range shorts {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("failed deleting short url `%s`: %w", v.ShortURL, err)
		}
	}
	return nil
}
