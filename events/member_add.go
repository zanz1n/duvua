package events

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/welcome"
	"github.com/zanz1n/duvua/pkg/pb/davinci"
)

type MemberAddEvent struct {
	r welcome.WelcomeRepository
	c davinci.DavinciClient
}

func NewMemberAddEvent(
	r welcome.WelcomeRepository,
	c davinci.DavinciClient,
) *MemberAddEvent {
	return &MemberAddEvent{r: r, c: c}
}

func (e *MemberAddEvent) Trigger(s *discordgo.Session, member *discordgo.Member) error {
	w, err := e.r.GetById(member.GuildID)

	if err != nil {
		return err
	} else if w == nil {
		return errors.New("a mensagem de boas vindas está desabilitada")
	} else if !w.Enabled {
		slog.Info(
			"Welcome message not sent: guild welcome disabled",
			"guild_id", member.GuildID,
		)
		return errors.New("a mensagem de boas vindas está desabilitada")
	} else if w.ChannelId == nil {
		slog.Info(
			"Welcome message not sent: enabled, but no channel set",
			"guild_id", member.GuildID,
		)
		return errors.New("nenhum canal de texto foi configurado")
	}

	msg := w.Message
	if strings.Contains(w.Message, "{{GUILD}}") {
		guild, err := s.Guild(member.GuildID)
		if err != nil {
			return err
		}

		msg = strings.ReplaceAll(msg, "{{GUILD}}", guild.Name)
	}

	var data discordgo.MessageSend

	switch w.Kind {
	case welcome.WelcomeTypeMessage:
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
	case welcome.WelcomeTypeImage:
		return e.handleImage(member, *w.ChannelId, msg)

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

func (e *MemberAddEvent) handleImage(member *discordgo.Member, channelId, msg string) error {
	var username string
	if member.User.GlobalName != "" {
		username = member.User.GlobalName
	} else {
		username = member.User.Username
	}

	var avatarUrl string
	if member.Avatar != "" {
		avatarUrl = discordgo.EndpointGuildMemberAvatar(member.GuildID, member.User.ID, member.Avatar)
	} else if member.User.Avatar != "" {
		avatarUrl = discordgo.EndpointUserAvatar(member.User.ID, member.User.Avatar)
	}

	disc := member.User.Discriminator
	if disc != "" && disc != "0" && disc != "0000" {
		username += "#" + member.User.Discriminator
	}
	msg = strings.ReplaceAll(msg, "{{USER}}", username)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := e.c.SendWelcome(ctx, &davinci.WelcomeRequest{
		Username:     username,
		ImageUrl:     avatarUrl,
		GreetingText: msg,
		Data: &davinci.ImageSendData{
			ChannelId: cuint64(channelId),
			Message:   "<@" + member.User.ID + ">",
			FileName:  fmt.Sprintf("%s-%s-welcome", member.GuildID, member.User.ID),
		},
	})
	return err
}

func cuint64(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 0)
	return v
}
