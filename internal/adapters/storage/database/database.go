package database

import (
	"context"
	"errors"
	"fmt"

	"database/sql"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/playmixer/short-link/internal/adapters/models"
	"github.com/playmixer/short-link/internal/adapters/shortnererror"
)

type Store struct {
	conn *sql.DB
}

func Conn(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed open database: %w", err)
	}

	return db, nil
}

func initTables(db *sql.DB) error {
	exec := `CREATE TABLE IF NOT EXISTS public.short_link (
		id int GENERATED ALWAYS AS IDENTITY NOT NULL,
		original_url varchar NOT NULL,
		short_url varchar NOT NULL,
		CONSTRAINT short_url_pk PRIMARY KEY (short_url)
	);
	CREATE UNIQUE INDEX IF NOT EXISTS short_link_original_url_idx ON public.short_link USING btree (original_url);
	CREATE UNIQUE INDEX IF NOT EXISTS short_link_short_url_idx ON public.short_link USING btree (short_url);`

	_, err := db.Exec(exec)
	if err != nil {
		return fmt.Errorf("error initialize tables: %w", err)
	}

	return nil
}

func New(cfg *Config) (*Store, error) {
	var err error
	s := &Store{}
	s.conn, err = Conn(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed connect to database: %w", err)
	}
	err = initTables(s.conn)
	if err != nil {
		return nil, fmt.Errorf("failed initialize tables: %w", err)
	}

	return s, nil
}

func (s *Store) Set(ctx context.Context, key, value string) error {
	_, err := s.conn.ExecContext(ctx, "insert into short_link (short_url, original_url) values ($1, $2)", key, value)
	var sqlError *pgconn.PgError
	if err != nil && errors.As(err, &sqlError) && pgerrcode.IsIntegrityConstraintViolation(sqlError.Code) {
		return fmt.Errorf("pgerror: %w: %w", shortnererror.ErrNotUnique, err)
	}
	if err != nil {
		return fmt.Errorf("failed setting short url: %w", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, key string) (string, error) {
	row := s.conn.QueryRowContext(ctx, "select original_url from short_link where short_url =$1", key)
	var value string
	err := row.Scan(&value)
	if err != nil {
		return "", fmt.Errorf("failed scan url %w", err)
	}
	if err := row.Err(); err != nil {
		return "", fmt.Errorf("query row error: %w", err)
	}
	return value, nil
}

func (s *Store) SetBatch(ctx context.Context, batch []models.ShortLink) error {
	tx, err := s.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmn, err := tx.PrepareContext(ctx, "insert into short_link (short_url, original_url) values ($1, $2)")
	if err != nil {
		return fmt.Errorf("cannot prepare sql query: %w", err)
	}
	defer func() { _ = stmn.Close() }()
	for _, query := range batch {
		_, err := stmn.ExecContext(ctx, query.ShortURL, query.OriginalURL)
		if err != nil {
			return fmt.Errorf("failed insert: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed commit insert: %w", err)
	}

	return nil
}

func (s *Store) GetByOriginal(ctx context.Context, original string) (string, error) {
	row := s.conn.QueryRowContext(ctx, "select short_url from short_link where original_url =$1", original)
	var value string
	err := row.Scan(&value)
	if err != nil {
		return "", fmt.Errorf("failed scan url %w", err)
	}
	if err := row.Err(); err != nil {
		return "", fmt.Errorf("query row error: %w", err)
	}
	return value, nil
}
