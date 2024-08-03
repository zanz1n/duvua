package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	gomigrate "github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/joho/godotenv"
	"github.com/zanz1n/duvua-bot/config"
	"github.com/zanz1n/duvua-bot/internal/commands"
	"github.com/zanz1n/duvua-bot/internal/events"
	"github.com/zanz1n/duvua-bot/internal/logger"
	"github.com/zanz1n/duvua-bot/internal/manager"
	"github.com/zanz1n/duvua-bot/internal/utils"
	"github.com/zanz1n/duvua-bot/internal/welcome"
	embedsql "github.com/zanz1n/duvua-bot/sql"
)

var (
	migrate = flag.Bool("migrate", false, "Migrates the database")
	debug   = flag.Bool("debug", false, "Enables debug logs")
)

var endCh chan os.Signal

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func init() {
	flag.Parse()
}

func init() {
	godotenv.Load()

	config := config.GetConfig()
	if *debug {
		config.LogLevel = slog.LevelDebug
	}

	slog.SetLogLoggerLevel(config.LogLevel)

	endCh = make(chan os.Signal, 1)
	signal.Notify(endCh, syscall.SIGINT, syscall.SIGTERM)
}

func main() {
	cfg := config.GetConfig()

	s, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		log.Fatalln("Failed to create discord session:", err)
	}

	s.Identify.Intents = discordgo.IntentsGuildMembers

	s.LogLevel = logger.SlogLevelToDiscordgo(cfg.LogLevel + 4)

	db := connectToPostgres()
	defer func() {
		start := time.Now()
		if e := db.Close(); e != nil {
			slog.Error(
				"Failed to close postgres client",
				"took", time.Since(start),
				"error", e,
			)
		} else {
			slog.Info("Closed postgres client", "took", time.Since(start))
		}
	}()

	if *migrate {
		if err = execMigration(db); err != nil {
			log.Fatalln("Failed to run migrations:", err)
		}
	}

	welcomeRepo := welcome.NewPostgresWelcomeRepository(db)

	m := manager.NewManager()

	m.Add(commands.NewHelpCommand(m))
	m.Add(commands.NewAvatarCommand())
	m.Add(commands.NewClearCommand())
	m.Add(commands.NewWelcomeCommand(welcomeRepo))

	m.AutoHandle(s)
	s.AddHandler(events.NewReadyEvent(m).Handle)
	s.AddHandler(events.NewMemberAddEvent(welcomeRepo).Handle)

	if err = s.Open(); err != nil {
		log.Fatalln("Failed to open discord session:", err)
	}
	defer func() {
		start := time.Now()
		if e := s.Close(); e != nil {
			slog.Error(
				"Failed to close discordgo session",
				"took", time.Since(start),
				"error", e,
			)
		} else {
			slog.Info("Closed discordgo session", "took", time.Since(start))
		}
	}()

	utils.SetStatus(s, utils.StatusTypeStarting)

	<-endCh
	log.Println("Closing bot ...")
	utils.SetStatus(s, utils.StatusTypeStopping)
}

func connectToPostgres() *sql.DB {
	cfg := config.GetConfig()

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
	cfg := config.GetConfig()

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
