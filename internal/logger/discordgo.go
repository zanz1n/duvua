package logger

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

func SlogLevelToDiscordgo(level slog.Level) int {
	switch level {
	case slog.LevelDebug:
		return discordgo.LogDebug
	case slog.LevelInfo:
		return discordgo.LogInformational
	case slog.LevelWarn:
		return discordgo.LogWarning
	case slog.LevelError:
		return discordgo.LogError
	default:
		return discordgo.LogInformational
	}
}
