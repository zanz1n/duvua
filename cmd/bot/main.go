package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
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
	"github.com/zanz1n/duvua-bot/internal/anime"
	configcmds "github.com/zanz1n/duvua-bot/internal/commands/config"
	funcmds "github.com/zanz1n/duvua-bot/internal/commands/fun"
	infocmds "github.com/zanz1n/duvua-bot/internal/commands/info"
	modcmds "github.com/zanz1n/duvua-bot/internal/commands/moderation"
	musiccmds "github.com/zanz1n/duvua-bot/internal/commands/music"
	ticketcmds "github.com/zanz1n/duvua-bot/internal/commands/ticket"
	"github.com/zanz1n/duvua-bot/internal/events"
	"github.com/zanz1n/duvua-bot/internal/lang"
	"github.com/zanz1n/duvua-bot/internal/logger"
	"github.com/zanz1n/duvua-bot/internal/manager"
	"github.com/zanz1n/duvua-bot/internal/music"
	"github.com/zanz1n/duvua-bot/internal/ticket"
	"github.com/zanz1n/duvua-bot/internal/utils"
	"github.com/zanz1n/duvua-bot/internal/welcome"
	"github.com/zanz1n/duvua-bot/pkg/player"
	embedsql "github.com/zanz1n/duvua-bot/sql"
	staticembed "github.com/zanz1n/duvua-bot/static"
)

const DuvuaBanner = `
 ____                          ____        _
|  _ \ _   ___   ___   _  __ _| __ )  ___ | |_
| | | | | | \ \ / / | | |/ _` + "`" + ` |  _ \ / _ \| __|
| |_| | |_| |\ V /| |_| | (_| | |_) | (_) | |_
|____/ \__,_| \_/  \__,_|\__,_|____/ \___/ \__|

Copyright © 2022 - %d Izan Rodrigues

Version: %s
     GO: %s
 Source: https://github.com/zanz1n/duvua-bot
License: https://github.com/zanz1n/duvua-bot/blob/main/LICENSE

This software is made available under the terms of the AGPL-3.0 license.

`

var (
	migrate  = flag.Bool("migrate", false, "Migrates the database")
	debug    = flag.Bool("debug", false, "Enables debug logs")
	jsonLogs = flag.Bool("json-logs", false, "Enables json logs")
	noBanner = flag.Bool("no-banner", false, "Disables the figlet banner")
)

var endCh chan os.Signal

func init() {
	flag.Parse()
	if !*jsonLogs && !*noBanner {
		fmt.Printf(
			DuvuaBanner[1:],
			time.Now().Year(),
			config.Version,
			runtime.Version(),
		)
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	if *jsonLogs {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	}
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

	s.Identify.Intents = discordgo.IntentsGuildMembers |
		discordgo.IntentsGuilds |
		discordgo.IntentGuildVoiceStates

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

	welcomeGen := welcomeImageGenerator()
	welcomeRepo := welcome.NewPostgresWelcomeRepository(db)
	welcomeEvt := events.NewMemberAddEvent(welcomeRepo, welcomeGen)

	ticketRepository := ticket.NewPgTicketRepository(db)
	ticketConfigRepository := ticket.NewPgTicketConfigRepository(db)

	musicRepository := music.NewPgMusicConfigRepository(db)
	musicClient := player.NewHttpClient(nil, cfg.Player.ApiURL, cfg.Player.Password)

	animeApi := anime.NewAnimeApi(nil)
	translator := lang.NewGoogleTranslatorApi(nil)

	m := manager.NewManager()

	m.Add(configcmds.NewWelcomeCommand(welcomeRepo, welcomeEvt))

	m.Add(funcmds.NewAvatarCommand())
	m.Add(funcmds.NewCloneCommand())
	m.Add(funcmds.NewShipCommand())

	m.Add(infocmds.NewFactsCommand())
	m.Add(infocmds.NewHelpCommand(m))
	m.Add(infocmds.NewPingCommand())
	m.Add(infocmds.NewAnimeCommand(animeApi, translator))

	m.Add(modcmds.NewClearCommand())

	m.Add(ticketcmds.NewTicketAdminCommand(ticketRepository, ticketConfigRepository))
	m.Add(ticketcmds.NewTicketCommand(ticketRepository, ticketConfigRepository))

	m.Add(musiccmds.NewMusicAdminCommand(musicRepository))
	m.Add(musiccmds.NewPlayCommand(musicRepository, musicClient))
	m.Add(musiccmds.NewSkipCommand(musicRepository, musicClient))
	m.Add(musiccmds.NewStopCommand(musicRepository, musicClient))
	m.Add(musiccmds.NewQueueCommand(musicRepository, musicClient))
	m.Add(musiccmds.NewLoopCommand(musicRepository, musicClient))

	m.AutoHandle(s)
	s.AddHandlerOnce(events.NewReadyEvent(m).Handle)
	s.AddHandler(welcomeEvt.Handle)
	s.AddHandler(events.NewChannelDeleteEvent(ticketRepository).Handle)

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

	sig := <-endCh
	log.Printf("Received signal %s: closing bot ...\n", sig.String())
	utils.SetStatus(s, utils.StatusTypeStopping)
}

func welcomeImageGenerator() *welcome.ImageGenerator {
	cfg := config.GetConfig()

	template, err := welcome.LoadTemplate(staticembed.Assets, "welcomer.png")
	if err != nil {
		log.Fatalln("Failed to load welcomer image template:", err)
	}

	font, err := welcome.LoadFont(staticembed.Assets, "jetbrains-mono.ttf")
	if err != nil {
		log.Fatalln("Failed to load welcomer image font:", err)
	}

	return welcome.NewImageGenerator(
		template,
		font,
		cfg.Welcomer.ImageQuality,
	)
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
