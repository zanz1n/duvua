package config

import (
	"fmt"
	"log/slog"
)

type Config struct {
	LogLevel slog.Level     `env:"LOG_LEVEL, default=0"`
	Discord  DiscordConfig  `env:", prefix=DISCORD_"`
	Postgres PostgresConfig `env:", prefix=POSTGRES_"`
	Welcomer WelcomerConfig `env:", prefix=WELCOMER_"`
	Player   PlayerConfig   `env:", prefix=PLAYER_"`
	Spotify  SpotifyConfig  `env:", prefix=SPOTIFY_"`
}

type DiscordConfig struct {
	Token string `env:"TOKEN, required"`
	// Nullable
	Guild *string `env:"GUILD, noinit"`
}

type PostgresConfig struct {
	Host string `env:"HOST, required"`
	Port uint16 `env:"PORT, default=5432"`

	Username string `env:"USER, required"`
	Password string `env:"PASSWORD, required"`
	SslMode  string `env:"SSL_MODE, default=disable"`

	Database string `env:"DB, required"`

	MaxConns int32 `env:"MAX_CONNS, default=32"`
	MinConns int32 `env:"MIN_CONNS, default=3"`
}

type WelcomerConfig struct {
	ImageQuality float32 `env:"IMAGE_QUALITY, default=80.0"`
}

type PlayerConfig struct {
	ApiURL     string `env:"URL, required"`
	ListenPort uint16 `env:"LISTEN_PORT, default=8080"`
	Password   string `env:"PASSWORD"`
	FFmpegExec string `env:"FFMPEG_EXEC"`
}

type SpotifyConfig struct {
	ClientId     string `env:"CLIENT_ID, required"`
	ClientSecret string `env:"CLIENT_SECRET, required"`
}

func (pc *PostgresConfig) IntoUri() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%v/%s?sslmode=%s",
		pc.Username,
		pc.Password,
		pc.Host,
		pc.Port,
		pc.Database,
		pc.SslMode,
	)
}
