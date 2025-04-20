package config

import (
	"fmt"
)

type DiscordConfig struct {
	Token string `env:"TOKEN, required"`
	Guild string `env:"GUILD"`
}

type PostgresConfig struct {
	Host string `env:"HOST, required"`
	Port uint16 `env:"PORT, default=5432"`

	Username string `env:"USER, required"`
	Password string `env:"PASSWORD, required"`
	SslMode  string `env:"SSL_MODE, default=disable"`

	Database string `env:"DB, required"`

	MaxConns int `env:"MAX_CONNS, default=32"`
	MinConns int `env:"MIN_CONNS, default=1"`
}

type WelcomerConfig struct {
	ApiURL       string  `env:"URL, required"`
	ListenPort   uint16  `env:"LISTEN_PORT, default=8080"`
	Password     string  `env:"PASSWORD"`
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
