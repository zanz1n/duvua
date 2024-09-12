package player

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/utils"
	"github.com/zanz1n/duvua/pkg/player"
)

type PlayerMessenger struct {
	s *discordgo.Session
}

func (m *PlayerMessenger) OnTrackStart(p *GuildPlayer, t *player.Track) {
	cid := p.GetMessageChannel()

	go func() {
		start := time.Now()

		if err := m.onTrackStart(cid, t); err != nil {
			slog.Error(
				"Messenger: Failed to send on-track-start message",
				"guild_id", p.GuildId,
				"took", time.Since(start).Round(time.Millisecond),
				"error", err,
			)
		} else {
			slog.Info(
				"Messenger: Sent on-track-start message",
				"guild_id", p.GuildId,
				"took", time.Since(start).Round(time.Millisecond),
			)
		}
	}()
}

func (m *PlayerMessenger) onTrackStart(cid uint64, t *player.Track) error {
	if cid == 0 {
		return errors.Unexpected("no text channel")
	}

	desc := fmt.Sprintf(
		"Tocando agora **[%s](%s)**\n\n**Duração: [%s]**",
		t.Data.Name,
		t.Data.URL,
		utils.FmtDuration(t.Data.Duration),
	)

	loopCustomId := "loop/on"
	if t.State != nil && t.State.Looping {
		desc = "**[Loop]** " + desc
		loopCustomId = "loop/off"
	}

	message := discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Description: desc,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: t.Data.Thumbnail,
			},
		}},
		Components: []discordgo.MessageComponent{discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Pular",
					Emoji:    emoji("⏭️"),
					Style:    discordgo.SecondaryButton,
					CustomID: "skip",
				},
				discordgo.Button{
					Label:    "Parar",
					Emoji:    emoji("⏹️"),
					Style:    discordgo.DangerButton,
					CustomID: "stop",
				},
				discordgo.Button{
					Label:    "Pause",
					Emoji:    emoji("⏸️"),
					Style:    discordgo.PrimaryButton,
					CustomID: "pause",
				},
				discordgo.Button{
					Label:    "Loop",
					Emoji:    emoji("🔁"),
					Style:    discordgo.SuccessButton,
					CustomID: loopCustomId,
				},
			},
		}},
	}

	return m.sendMessage(cid, &message)
}

func (m *PlayerMessenger) OnQueueEnd(p *GuildPlayer) {
	cid := p.GetMessageChannel()

	go func() {
		start := time.Now()

		if err := m.onQueueEnd(cid); err != nil {
			slog.Error(
				"Messenger: Failed to send on-queue-end message",
				"guild_id", p.GuildId,
				"took", time.Since(start).Round(time.Millisecond),
				"error", err,
			)
		} else {
			slog.Info(
				"Messenger: Sent on-queue-end message",
				"guild_id", p.GuildId,
				"took", time.Since(start).Round(time.Millisecond),
			)
		}
	}()
}

func (m *PlayerMessenger) onQueueEnd(cid uint64) error {
	if cid == 0 {
		return errors.Unexpected("no text channel")
	}

	return m.sendMessage(cid, &discordgo.MessageSend{
		Content: "**A fila (playlist) terminou!**",
	})
}

func (m *PlayerMessenger) OnTrackFailed(p *GuildPlayer, t *player.Track) {
	cid := p.GetMessageChannel()

	go func() {
		start := time.Now()

		if err := m.onTrackFailed(cid, t); err != nil {
			slog.Error(
				"Messenger: Failed to send on-track-failed message",
				"guild_id", p.GuildId,
				"took", time.Since(start).Round(time.Millisecond),
				"error", err,
			)
		} else {
			slog.Info(
				"Messenger: Sent on-track-failed message",
				"guild_id", p.GuildId,
				"took", time.Since(start).Round(time.Millisecond),
			)
		}
	}()
}

func (m *PlayerMessenger) onTrackFailed(cid uint64, t *player.Track) error {
	if cid == 0 {
		return errors.Unexpected("no text channel")
	}

	return m.sendMessage(cid, &discordgo.MessageSend{
		Content: fmt.Sprintf(
			"Não foi possível tocar a música **[%s](%s)**",
			t.Data.Name,
			t.Data.URL,
		),
	})
}

func (m *PlayerMessenger) sendMessage(cid uint64, data *discordgo.MessageSend) error {
	_, err := m.s.ChannelMessageSendComplex(
		strconv.FormatUint(cid, 10),
		data,
	)
	if err != nil {
		return errors.Unexpected(
			"failed to send message on text channel: " + err.Error(),
		)
	}
	return nil
}

func emoji(name string) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: name}
}
