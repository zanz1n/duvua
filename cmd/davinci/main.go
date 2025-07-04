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

	"github.com/zanz1n/duvua/config"
	"github.com/zanz1n/duvua/internal/davinci"
	"github.com/zanz1n/duvua/internal/utils/grpcutils"
	davincipb "github.com/zanz1n/duvua/pkg/pb/davinci"
	staticembed "github.com/zanz1n/duvua/static"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const DuvuaBanner = `
 ____                          ____      __     ___            _
|  _ \ _   ___   ___   _  __ _|  _ \  __ \ \   / (_)_ __   ___(_)
| | | | | | \ \ / / | | |/ _` + "` | | | |/ _`" + ` \ \ / /| | '_ \ / __| |
| |_| | |_| |\ V /| |_| | (_| | |_| | (_| |\ V / | | | | | (__| |
|____/ \__,_| \_/  \__,_|\__,_|____/ \__,_| \_/  |_|_| |_|\___|_|

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
			"Running davinci",
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
	cfg := GetConfig()
	if *debug {
		cfg.LogLevel = slog.LevelDebug
	}

	slog.SetLogLoggerLevel(cfg.LogLevel)

	endCh = make(chan os.Signal, 1)
	signal.Notify(endCh, syscall.SIGINT, syscall.SIGTERM)
}

func main() {
	cfg := GetConfig()

	template, err := davinci.LoadTemplate(staticembed.Assets, "welcomer.png")
	if err != nil {
		log.Fatalln("Failed to load welcomer image template:", err)
	}

	font, err := davinci.LoadFont(staticembed.Assets, "jetbrains-mono.ttf")
	if err != nil {
		log.Fatalln("Failed to load welcomer image font:", err)
	}

	welcomeGen := davinci.NewWelcomeGenerator(
		template,
		font,
		cfg.Welcomer.ImageQuality,
	)

	server := davinci.NewGrpcServer("Bot "+cfg.Discord.Token, welcomeGen, nil)

	listenAddr := fmt.Sprintf("0.0.0.0:%d", cfg.Welcomer.ListenPort)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen grpc on `%s`: %s\n", listenAddr, err)
	}

	slog.Info("GRPC: Listening for grpc connections", "addr", listenAddr)

	passwd := cfg.Welcomer.Password
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(grpcutils.AllUnaryServerInterceptors(passwd)...),
		grpc.ChainStreamInterceptor(grpcutils.AllStreamServerInterceptors(passwd)...),
	)

	davincipb.RegisterDavinciServer(grpcServer, server)
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

	sig := <-endCh
	log.Printf("Received signal %s: closing image generator ...\n", sig.String())
}
