package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/joho/godotenv"
	"github.com/zanz1n/duvua-bot/config"
	"github.com/zanz1n/duvua-bot/internal/commands"
	"github.com/zanz1n/duvua-bot/internal/events"
	"github.com/zanz1n/duvua-bot/internal/logger"
	"github.com/zanz1n/duvua-bot/internal/manager"
)

var endCh chan os.Signal

func init() {
	godotenv.Load()

	config := config.GetConfig()

	slog.SetLogLoggerLevel(config.LogLevel)

	endCh = make(chan os.Signal, 1)
	signal.Notify(endCh, syscall.SIGINT, syscall.SIGTERM)
}

func main() {
	config := config.GetConfig()

	s, err := discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		log.Fatalln("Failed to create discord session:", err)
	}

	s.LogLevel = logger.SlogLevelToDiscordgo(config.LogLevel)

	pgxConfig, err := pgxpool.ParseConfig(config.Postgres.IntoUri())
	if err != nil {
		log.Fatalln("Failed to parse postgres config:", err)
	}

	pgxConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   logger.NewPgxLogger(slog.Default()),
		LogLevel: logger.SlogLevelToPgx(config.LogLevel),
	}
	pgxConfig.MaxConns = config.Postgres.MaxConns
	pgxConfig.MinConns = config.Postgres.MinConns

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		log.Fatalln("Failed to connect to postgres:", err)
	}

	m := manager.NewManager()

	m.Add(commands.NewHelpCommand())

	m.AutoHandle(s)
	s.AddHandler(events.NewReadyEvent(m).Handle)

	if err = s.Open(); err != nil {
		log.Fatalln("Failed to open discord session:", err)
	}

	<-endCh
	log.Println("Closing bot ...")

	pool.Close()
	s.Close()
}
