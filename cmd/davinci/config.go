package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
	"github.com/zanz1n/duvua/config"
)

type Config struct {
	LogLevel slog.Level            `env:"LOG_LEVEL, default=0"`
	Discord  config.DiscordConfig  `env:", prefix=DISCORD_"`
	Welcomer config.WelcomerConfig `env:", prefix=WELCOMER_"`
}

var configInstance *Config = nil

func GetConfig() *Config {
	if configInstance == nil {
		if err := InitConfig(); err != nil {
			log.Fatalln("Failed to init bot configuration:", err)
		}
	}

	return configInstance
}

func InitConfig() error {
	godotenv.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	configInstance = &Config{}

	return envconfig.Process(ctx, configInstance)
}
