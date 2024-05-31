package database

import (
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

func New(cfg *Config) (*Store, error) {
	var err error
	s := &Store{}
	s.conn, err = Conn(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed connect to database: %w", err)
	}

	return s, nil
}

func (s *Store) Set(key, value string) error {
	return nil
}

func (s *Store) Get(key string) (string, error) {
	return "", nil
}
