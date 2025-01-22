package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/config"
	"github.com/zanz1n/duvua/internal/grpcutils"
	"github.com/zanz1n/duvua/internal/logger"
	"github.com/zanz1n/duvua/internal/player"
	"github.com/zanz1n/duvua/internal/player/platform"
	playerpb "github.com/zanz1n/duvua/pkg/pb/player"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
 Source: https://github.com/zanz1n/duvua
License: https://github.com/zanz1n/duvua/blob/main/LICENSE

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
	} else {
		slog.Info(
			"Running player",
			"version", config.Version,
			"go", runtime.Version(),
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
	cfg := config.GetConfig()
	if *debug {
		cfg.LogLevel = slog.LevelDebug
	}

	slog.SetLogLoggerLevel(cfg.LogLevel)

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

	s.LogLevel = logger.SlogLevelToDiscordgo(cfg.LogLevel + 4)

	ytFetcher := platform.NewYoutube(nil, 1)
	spotifyFetcher, err := platform.NewSpotify(
		cfg.Spotify.ClientId,
		cfg.Spotify.ClientSecret,
		ytFetcher,
	)
	if err != nil {
		log.Fatalln(err)
	}

	fetcher := platform.NewFetcher(ytFetcher, spotifyFetcher)
	manager := player.NewPlayerManager(s, fetcher)

	defer manager.Close()

	server := player.NewGrpcServer(manager, fetcher)

	listenAddr := fmt.Sprintf("0.0.0.0:%d", cfg.Player.ListenPort)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen grpc on `%s`: %s\n", listenAddr, err)
	}

	slog.Info("GRPC: Listening for grpc connections", "addr", listenAddr)

	passwd := cfg.Player.Password
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(grpcutils.AllUnaryServerInterceptors(passwd)...),
		grpc.ChainStreamInterceptor(grpcutils.AllStreamServerInterceptors(passwd)...),
	)

	playerpb.RegisterPlayerServer(grpcServer, server)
	reflection.Register(grpcServer)

	go grpcServer.Serve(ln)
	defer func() {
		start := time.Now()
		grpcServer.GracefulStop()
		slog.Info(
			"Closed grpc server gracefully",
			"took", time.Since(start).Round(time.Millisecond),
		)
	}()

	readyStart := time.Now()
	s.AddHandler(func(s *discordgo.Session, ready *discordgo.Ready) {
		slog.Info(
			"Discord session ready",
			"username", s.State.User.Username+"#"+s.State.User.Discriminator,
			"took", time.Since(readyStart).Round(time.Millisecond),
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
			slog.Info(
				"Closed discordgo session",
				"took", time.Since(start).Round(time.Millisecond),
			)
		}
	}()

	sig := <-endCh
	log.Printf("Received signal %s: closing player ...\n", sig.String())
}
