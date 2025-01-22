package events

import (
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/internal/manager"
	"github.com/zanz1n/duvua/internal/utils"
)

type ReadyEvent struct {
	m       *manager.Manager
	start   time.Time
	guildId string
}

func NewReadyEvent(m *manager.Manager, guildId string) *ReadyEvent {
	return &ReadyEvent{m: m, start: time.Now()}
}

func (re *ReadyEvent) Handle(s *discordgo.Session, ready *discordgo.Ready) {
	slog.Info(
		"Discord session ready",
		"username", s.State.User.Username+"#"+s.State.User.Discriminator,
		"took", time.Since(re.start),
	)

	re.m.PostCommands(s, re.guildId)

	utils.SetStatus(s, utils.StatusTypeIdle)
}
