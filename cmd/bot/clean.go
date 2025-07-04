package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sethvargo/go-envconfig"
	"github.com/zanz1n/duvua/config"
	"github.com/zanz1n/duvua/internal/logger"
)

func runClean() {
	start := time.Now()

	var cfg struct {
		LogLevel slog.Level           `env:"LOG_LEVEL, default=info"`
		Discord  config.DiscordConfig `env:", prefix=DISCORD_"`
	}

	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		log.Fatalln("Failed to init configuration:", err)
	}

	s, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		log.Fatalln("Failed to create discord session:", err)
	}

	s.LogLevel = logger.SlogLevelToDiscordgo(cfg.LogLevel + 4)

	if err := s.Open(); err != nil {
		log.Fatalln("Failed to open discord session:", err)
	}
	defer s.Close()

	slog.Info("Cleaning commands",
		"application_id", s.State.Application.ID,
		"user_id", s.State.User.ID,
		"guild_id", cfg.Discord.Guild,
	)

	_, err = s.ApplicationCommandBulkOverwrite(
		s.State.Application.ID,
		cfg.Discord.Guild,
		[]*discordgo.ApplicationCommand{},
	)
	if err != nil {
		log.Fatalln("Failed to clean commands:", err)
	}

	slog.Info(
		"Successfully cleaned commands",
		"took", time.Since(start).Round(time.Millisecond),
	)
}
