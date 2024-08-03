package events

import (
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/internal/welcome"
)

type MemberAddEvent struct {
	r welcome.WelcomeRepository
}

func NewMemberAddEvent(r welcome.WelcomeRepository) *MemberAddEvent {
	return &MemberAddEvent{r: r}
}

func (e *MemberAddEvent) Trigger(s *discordgo.Session, member *discordgo.Member) error {
	w, err := e.r.GetById(member.GuildID)

	if err != nil {
		return err
	} else if w == nil {
		return nil
	} else if !w.Enabled {
		slog.Info(
			"Guild welcome message is disabled",
			"guild_id", member.GuildID,
		)
		return errors.New("a mensagem de boas vindas está desabilitada")
	} else if w.ChannelId == nil {
		slog.Info(
			"Guild welcome is enabled, but no channel was set",
			"guild_id", member.GuildID,
		)
		return errors.New("a mensagem de boas vindas")
	}

	msg := w.Message
	if strings.Contains(w.Message, "{{GUILD}}") {
		guild, err := s.State.Guild(member.GuildID)
		if err != nil {
			guild, err = s.Guild(member.GuildID)
			if err != nil {
				return err
			}
		}

		msg = strings.ReplaceAll(msg, "{{GUILD}}", guild.Name)
	}

	var data discordgo.MessageSend

	switch w.Kind {
	case welcome.WelcomeTypeMessage, welcome.WelcomeTypeImage:
		msg = strings.ReplaceAll(msg, "{{USER}}", "<@"+member.User.ID+">")

		data = discordgo.MessageSend{
			Content: msg,
		}
	case welcome.WelcomeTypeEmbed:
		msg = strings.ReplaceAll(msg, "{{USER}}", "<@"+member.User.ID+">")

		data = discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{{
				Type:        discordgo.EmbedTypeArticle,
				Description: msg,
			}},
		}
	default:
		slog.Error(
			"Welcome has an invalid type",
			"type", w.Kind,
			"guild_id", member.GuildID,
		)
		return nil
	}

	_, err = s.ChannelMessageSendComplex(*w.ChannelId, &data)
	if err != nil {
		slog.Error(
			"Failed to send welcome message on channel",
			"guild_id", member.GuildID,
			"channel_id", *w.ChannelId,
			"error", err,
		)
		return errors.Newf(
			"não foi possível enviar a mensagem no canal de texto <#%s>",
			*w.ChannelId,
		)
	}

	return nil
}

func (e *MemberAddEvent) Handle(s *discordgo.Session, member *discordgo.GuildMemberAdd) {
	start := time.Now()

	err := e.Trigger(s, member.Member)
	if err != nil {
		exp := false
		if e, ok := err.(errors.Expected); ok {
			exp = e.IsExpected()
		}
		if !exp {
			slog.Error(
				"Something went wrong handling guild-member-add event",
				"guild_id", member.GuildID,
				"user_id", member.User.ID,
				"took", time.Since(start),
				"error", err,
			)
		}
	} else {
		slog.Info(
			"Handled guild-member-add event",
			"guild_id", member.GuildID,
			"user_id", member.User.ID,
			"took", time.Since(start),
		)
	}
}
