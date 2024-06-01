package database

import (
	"context"
	"fmt"

	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
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
	CREATE INDEX IF NOT EXISTS short_url_short_url_idx ON public.short_link (short_url);`

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
