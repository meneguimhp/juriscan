package app

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

func openApplicationDB(cfg Config) (*sql.DB, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.DatabaseDriver)) {
	case "", "sqlite":
		return openSQLiteDB(cfg.DBPath)
	case "mysql":
		return openMySQLDB(cfg.DatabaseURL)
	default:
		return nil, fmt.Errorf("database: unsupported driver %q", cfg.DatabaseDriver)
	}
}

func openSQLiteDB(path string) (*sql.DB, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "juriscan.db"
	}

	dsn := path
	if strings.EqualFold(path, ":memory:") {
		dsn = "file::memory:?cache=shared"
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite ping: %w", err)
	}
	return db, nil
}

func openMySQLDB(dsn string) (*sql.DB, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, fmt.Errorf("mysql dsn required")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql open: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("mysql ping: %w", err)
	}
	return db, nil
}
