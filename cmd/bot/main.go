package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/commands"
	"github.com/zanz1n/duvua/config"
	"github.com/zanz1n/duvua/events"
	"github.com/zanz1n/duvua/internal/anime"
	"github.com/zanz1n/duvua/internal/lang"
	"github.com/zanz1n/duvua/internal/manager"
	"github.com/zanz1n/duvua/internal/music"
	"github.com/zanz1n/duvua/internal/ticket"
	"github.com/zanz1n/duvua/internal/utils"
	"github.com/zanz1n/duvua/internal/utils/logger"
	"github.com/zanz1n/duvua/internal/welcome"
	"github.com/zanz1n/duvua/pkg/pb/davinci"
	"github.com/zanz1n/duvua/pkg/pb/player"
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
 Source: https://github.com/zanz1n/duvua
License: https://github.com/zanz1n/duvua/blob/main/LICENSE

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
	} else {
		slog.Info(
			"Running bot",
			"version", config.Version,
			"go", runtime.Version(),
		)
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	if *jsonLogs {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	}
}

func init() {
	args := flag.Args()
	if len(args) == 0 {
		return
	} else if len(args) > 1 {
		log.Fatalln(
			"More than one argument provided:",
			strings.Join(args, ", "),
		)
	}
	arg := args[0]

	switch arg {
	case "run", "start":

	case "clean":
		runClean()
		defer os.Exit(0)

	default:
		log.Fatalln("Invalid argument:", arg)
	}
}

func init() {
	config := GetConfig()
	if *debug {
		config.LogLevel = slog.LevelDebug
	}

	slog.SetLogLoggerLevel(config.LogLevel)

	endCh = make(chan os.Signal, 1)
	signal.Notify(endCh, syscall.SIGINT, syscall.SIGTERM)
}

func main() {
	cfg := GetConfig()

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

	playerGrpc, playerCancel := connectToPlayerGrpc()
	defer playerCancel()

	davinciGrpc, davinciCancel := connectToDavinciGrpc()
	defer davinciCancel()

	welcomeRepo := welcome.NewPostgresWelcomeRepository(db)
	davinciClient := davinci.NewDavinciClient(davinciGrpc)
	welcomeEvt := events.NewMemberAddEvent(welcomeRepo, davinciClient)

	ticketRepository := ticket.NewPgTicketRepository(db)
	ticketConfigRepository := ticket.NewPgTicketConfigRepository(db)

	musicRepository := music.NewPgMusicConfigRepository(db)

	musicClient := player.NewPlayerClient(playerGrpc)

	animeApi := anime.NewAnimeApi(nil)
	translator := lang.NewGoogleTranslatorApi(nil)

	m := manager.NewManager()

	commands.Wire(m,
		welcomeRepo,
		welcomeEvt,
		animeApi,
		translator,
		ticketRepository,
		ticketConfigRepository,
		musicRepository,
		musicClient,
	)

	m.AutoHandle(s)
	s.AddHandlerOnce(events.NewReadyEvent(m, cfg.Discord.Guild).Handle)
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
