package database

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattes/migrate/source/file"
)

func runMigrations(dsn string) error {
	curPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed getting current path: %w", err)
	}
	migrationDir := strings.ReplaceAll(path.Join(curPath, "migrations"), string(os.PathSeparator), "/")
	sourceURL := "file://" + migrationDir
	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed apply migrations to DB: %w", err)
		}
	}

	return nil
}
