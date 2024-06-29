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

func New(ctx context.Context, cfg *Config) (*Store, error) {
	var err error

	if err = runMigrations(cfg.DSN); err != nil {
		return nil, fmt.Errorf("failed initialize tables: %w", err)
	}

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

	return s, nil
}

func (s *Store) Set(ctx context.Context, userID, short, original string) (output string, err error) {
	output, err = s.getByOriginal(ctx, userID, original)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return output, fmt.Errorf("failed select URL %s %w", original, err)
	}
	if output != "" {
		return output, fmt.Errorf("url `%s` is not unique: %w", original, storeerror.ErrNotUnique)
	}

	_, err = s.pool.Exec(
		ctx,
		"insert into short_link (short_url, original_url, user_id) values ($1, $2, $3)",
		short, original, userID,
	)
	if err != nil {
		var sqlError *pgconn.PgError
		if errors.As(err, &sqlError) && pgerrcode.UniqueViolation == sqlError.Code {
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
	for _, d := range data {
		short, err := s.getByOriginal(ctx, userID, d.OriginalURL)
		if err != nil && errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return []models.ShortLink{}, fmt.Errorf("failed getting URL `%s`: %w", d.OriginalURL, err)
		}
		if short != "" {
			return []models.ShortLink{{ShortURL: short, OriginalURL: d.OriginalURL}},
				fmt.Errorf("URL `%s` is not unique: %w", d.OriginalURL, storeerror.ErrNotUnique)
		}
	}

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
		"select short_url from short_link where original_url =$1 and user_id = $2 and is_deleted = false",
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
	rows, err := s.pool.Query(ctx,
		"select short_url, original_url from short_link where user_id = $1 and is_deleted = false", userID)
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

func (s *Store) HardDeleteURLs(ctx context.Context) error {
	sqlString := `delete from short_link where is_deleted = true`
	_, err := s.pool.Exec(ctx, sqlString)
	if err != nil {
		return fmt.Errorf("failed hard deleting URLs: %w", err)
	}
	return nil
}
