package events

import (
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/internal/ticket"
)

type ChannelDeleteEvent struct {
	r ticket.TicketRepository
}

func NewChannelDeleteEvent(r ticket.TicketRepository) *ChannelDeleteEvent {
	return &ChannelDeleteEvent{r: r}
}

func (e *ChannelDeleteEvent) Trigger(s *discordgo.Session, c *discordgo.Channel) error {
	t, err := e.r.DeleteByChannelId(c.ID)
	if err != nil {
		return err
	} else if t == nil {
		return nil
	}

	slog.Info(
		"Deleted ticket because it's channel has been deleted",
		"slug", t.Slug,
		"channel_id", t.ChannelId,
	)

	return nil
}

func (e *ChannelDeleteEvent) Handle(s *discordgo.Session, c *discordgo.ChannelDelete) {
	start := time.Now()

	err := e.Trigger(s, c.Channel)
	if err != nil {
		exp := false
		if e, ok := err.(errors.Expected); ok {
			exp = e.IsExpected()
		}
		if !exp {
			slog.Error(
				"Something went wrong handling channel-delete event",
				"guild_id", c.GuildID,
				"channel_id", c.ID,
				"took", time.Since(start),
				"error", err,
			)
		}
	} else {
		slog.Info(
			"Handled channel-delete event",
			"guild_id", c.GuildID,
			"channel_id", c.ID,
			"took", time.Since(start),
		)
	}
}
