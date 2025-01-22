package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"time"

	gomigrate "github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/zanz1n/duvua/internal/logger"
	embedsql "github.com/zanz1n/duvua/sql"
)

func connectToPostgres() *sql.DB {
	cfg := GetConfig()

	pgxConfig, err := pgxpool.ParseConfig(cfg.Postgres.IntoUri())
	if err != nil {
		log.Fatalln("Failed to parse postgres config:", err)
	}

	pgxConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   logger.NewPgxLogger(slog.Default()),
		LogLevel: logger.SlogLevelToPgx(cfg.LogLevel + 4),
	}
	pgxConfig.MaxConns = cfg.Postgres.MaxConns
	pgxConfig.MinConns = cfg.Postgres.MinConns

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbStart := time.Now()
	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		log.Fatalln("Failed to connect to postgres:", err)
	}

	slog.Info("Connected to database", "took", time.Since(dbStart))

	db := stdlib.OpenDBFromPool(pool)

	if err = db.Ping(); err != nil {
		log.Fatalln("Failed to connect to postgres:", err)
	}

	return db
}

func execMigration(db *sql.DB) error {
	start := time.Now()
	cfg := GetConfig()

	source, err := iofs.New(embedsql.Migrations, "migrations")
	if err != nil {
		return err
	}

	driver, err := migratepgx.WithInstance(db, &migratepgx.Config{
		DatabaseName:     cfg.Postgres.Database,
		StatementTimeout: 10 * time.Second,
	})
	if err != nil {
		return err
	}

	migrator, err := gomigrate.NewWithInstance("iofs", source, "pgx5", driver)
	if err != nil {
		return err
	}

	if err = migrator.Up(); err != nil {
		if err != gomigrate.ErrNoChange {
			return err
		}
	}

	slog.Info("Migrations were run", "took", time.Since(start))

	return nil
}
