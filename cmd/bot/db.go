package main

import (
	"database/sql"
	"log"
	"log/slog"
	"time"

	gomigrate "github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	embedsql "github.com/zanz1n/duvua/sql"
)

func connectToPostgres() *sql.DB {
	cfg := GetConfig()

	dbStart := time.Now()

	pgxConfig, err := pgx.ParseConfig(cfg.Postgres.IntoUri())
	if err != nil {
		log.Fatalln("Failed to parse postgres config:", err)
	}

	cfgId := stdlib.RegisterConnConfig(pgxConfig)

	db, err := sql.Open("pgx/v5", cfgId)
	if err != nil {
		log.Fatalln("Failed to connect to postgres:", err)
	}
	db.SetMaxIdleConns(cfg.Postgres.MaxConns)

	if cfg.Postgres.MinConns > 0 {
		if err = db.Ping(); err != nil {
			log.Fatalln("Failed to ping postgres database:", err)
		}
	}

	slog.Info("Connected to database", "took", time.Since(dbStart))

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
