package events

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/config"
	"github.com/zanz1n/duvua-bot/internal/manager"
)

type ReadyEvent struct {
	m *manager.Manager
}

func NewReadyEvent(m *manager.Manager) *ReadyEvent {
	return &ReadyEvent{m: m}
}

func (re *ReadyEvent) Handle(s *discordgo.Session, ready *discordgo.Ready) {
	slog.Info(
		"Logged to discord",
		"username",
		s.State.User.Username+"#"+s.State.User.Discriminator,
	)

	re.m.PostCommands(s, config.GetConfig().Discord.Guild)
}
