package main

import (
	"log/slog"

	"github.com/zanz1n/duvua/config"
	"github.com/zanz1n/duvua/internal/utils"
)

type Config struct {
	LogLevel slog.Level            `env:"LOG_LEVEL, default=0"`
	Discord  config.DiscordConfig  `env:", prefix=DISCORD_"`
	Postgres config.PostgresConfig `env:", prefix=POSTGRES_"`
	Welcomer config.WelcomerConfig `env:", prefix=WELCOMER_"`
	Player   config.PlayerConfig   `env:", prefix=PLAYER_"`
}

var configInstance = utils.NewLazyConfig[Config]()

func GetConfig() *Config {
	return configInstance.Get()
}

func InitConfig() error {
	return configInstance.Init()
}
