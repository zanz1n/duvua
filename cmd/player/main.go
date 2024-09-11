package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/config"
	"github.com/zanz1n/duvua-bot/internal/player"
)

const DuvuaBanner = `
 ____                          ____  _
|  _ \ _   ___   ___   _  __ _|  _ \| | __ _ _   _  ___ _ __
| | | | | | \ \ / / | | |/ _` + "` | |_) | |/ _`" + ` | | | |/ _ \ '__|
| |_| | |_| |\ V /| |_| | (_| |  __/| | (_| | |_| |  __/ |
|____/ \__,_| \_/  \__,_|\__,_|_|   |_|\__,_|\__, |\___|_|
                                             |___/

Copyright © 2022 - %d Izan Rodrigues

Version: %s
     GO: %s
 Source: https://github.com/zanz1n/duvua-bot
License: https://github.com/zanz1n/duvua-bot/blob/main/LICENSE

This software is made available under the terms of the AGPL-3.0 license.

`

var (
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

	s.Identify.Intents = discordgo.IntentGuildVoiceStates

	fetcher := player.NewTrackFetcher(player.NewYoutubeFetcher(nil, 1))
	manager := player.NewPlayerManager(s, fetcher)
	handler := player.NewHandler(manager, fetcher)

	mux := http.NewServeMux()

	player.NewHttpServer(handler, cfg.Player.Password).Route(mux)

	go func() {
		listenAddr := fmt.Sprintf("0.0.0.0:%d", cfg.Player.ListenPort)
		slog.Info("HTTP: Listening for http connections", "addr", listenAddr)

		if err := http.ListenAndServe(listenAddr, mux); err != nil {
			log.Fatalln("Failed to listen http:", err)
		}
	}()

	readyStart := time.Now()
	s.AddHandler(func(s *discordgo.Session, ready *discordgo.Ready) {
		slog.Info(
			"Discord session ready",
			"username", s.State.User.Username+"#"+s.State.User.Discriminator,
			"took", time.Since(readyStart),
		)
	})

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

	sig := <-endCh
	log.Printf("Received signal %s: closing bot ...\n", sig.String())
}
