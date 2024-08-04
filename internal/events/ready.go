package events

import (
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/config"
	"github.com/zanz1n/duvua-bot/internal/manager"
	"github.com/zanz1n/duvua-bot/internal/utils"
)

type ReadyEvent struct {
	m     *manager.Manager
	start time.Time
}

func NewReadyEvent(m *manager.Manager) *ReadyEvent {
	return &ReadyEvent{m: m, start: time.Now()}
}

func (re *ReadyEvent) Handle(s *discordgo.Session, ready *discordgo.Ready) {
	slog.Info(
		"Discord session ready",
		"username", s.State.User.Username+"#"+s.State.User.Discriminator,
		"took", time.Since(re.start),
	)

	re.m.PostCommands(s, config.GetConfig().Discord.Guild)

	utils.SetStatus(s, utils.StatusTypeIdle)
}
