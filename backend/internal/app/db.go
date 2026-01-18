package app

import (
	"backend/internal/infrastructure/repositories/postgres"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

const (
	migrationsPath = "./migrations"
	driverName     = "pgx"
	dbTypePostgres = "postgres"
)

func initDB(cfg postgres.Config) (*sqlx.DB, error) {
	dataSource := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s?&sslmode=disable&default_query_exec_mode=cache_describe&search_path=public",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Database,
	)

	db, err := sqlx.Open(driverName, dataSource)
	if err != nil {
		return nil, fmt.Errorf("create pool of connections to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err = runMigrations(db.DB); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	if err := goose.SetDialect(dbTypePostgres); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.Up(db, migrationsPath); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
