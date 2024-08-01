package config

import (
	"fmt"
	"log/slog"
)

type Config struct {
	LogLevel slog.Level     `env:"LOG_LEVEL, default=0"`
	Discord  DiscordConfig  `env:", prefix=DISCORD_"`
	Postgres PostgresConfig `env:", prefix=POSTGRES_"`
}

type DiscordConfig struct {
	Token string `env:"TOKEN, required"`
	// Nullable
	Guild *string `env:"GUILD, noinit"`
}

type PostgresConfig struct {
	Host string `env:"HOST, required"`
	Port uint16 `env:"PORT, default=5432"`

	Username string `env:"USERNAME, required"`
	Password string `env:"PASSWORD, required"`
	SslMode  string `env:"SSL_MODE, default=disable"`

	Database string `env:"DB, required"`

	MaxConns int32 `env:"MAX_CONNS, default=32"`
	MinConns int32 `env:"MIN_CONNS, default=3"`
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
