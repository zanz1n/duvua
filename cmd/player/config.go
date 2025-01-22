package main

import (
	"log/slog"

	"github.com/zanz1n/duvua/config"
	"github.com/zanz1n/duvua/internal/utils"
)

type Config struct {
	LogLevel slog.Level           `env:"LOG_LEVEL, default=info"`
	Discord  config.DiscordConfig `env:", prefix=DISCORD_"`
	Player   config.PlayerConfig  `env:", prefix=PLAYER_"`
	Spotify  config.SpotifyConfig `env:", prefix=SPOTIFY_"`
}

var configInstance = utils.NewLazyConfig[Config]()

func GetConfig() *Config {
	return configInstance.Get()
}

func InitConfig() error {
	return configInstance.Init()
}
