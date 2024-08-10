package events

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/internal/welcome"
	staticembed "github.com/zanz1n/duvua-bot/static"
)

type MemberAddEvent struct {
	r   welcome.WelcomeRepository
	gen *welcome.ImageGenerator
}

func NewMemberAddEvent(
	r welcome.WelcomeRepository,
	gen *welcome.ImageGenerator,
) *MemberAddEvent {
	return &MemberAddEvent{r: r, gen: gen}
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
		var username string
		if member.User.GlobalName != "" {
			username = member.User.GlobalName
		} else {
			username = member.User.Username
		}

		if member.User.Discriminator != "" {
			username += "#" + member.User.Discriminator
		}
		msg = strings.ReplaceAll(msg, "{{USER}}", username)

		img, err := e.generateImage(s, member, username, msg)
		if err != nil {
			return err
		}

		data = discordgo.MessageSend{
			Content: "<@" + member.User.ID + ">",
			Files: []*discordgo.File{{
				Name:        fmt.Sprintf("%s-%s-welcome.png", member.GuildID, member.User.ID),
				ContentType: "image/webp",
				Reader:      img,
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

func (e *MemberAddEvent) fetchAvatar(s *discordgo.Session, member *discordgo.Member) (io.ReadCloser, error) {
	var url string
	if member.Avatar != "" {
		url = discordgo.EndpointGuildMemberAvatar(member.GuildID, member.User.ID, member.Avatar)
	} else if member.User.Avatar != "" {
		url = discordgo.EndpointUserAvatar(member.User.ID, member.User.Avatar)
	} else {
		path := fmt.Sprintf("avatar/default-%v.png", rand.Intn(6))
		return staticembed.Assets.Open(path)
	}

	res, err := s.Client.Get(url)
	if err != nil {
		return nil, err
	}
	mt, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))

	if res.StatusCode != http.StatusOK {
		return nil, errors.Unexpected("fetch avatar: unexpected status code")
	} else if mt != "image/png" {
		return nil, errors.Unexpectedf("fetch avatar: unexpected image mime type `%s`", mt)
	}

	return res.Body, nil
}

func (e *MemberAddEvent) generateImage(
	s *discordgo.Session,
	member *discordgo.Member,
	username, msg string,
) (io.Reader, error) {
	av, err := e.fetchAvatar(s, member)
	if err != nil {
		return nil, err
	}
	defer av.Close()

	buf, err := e.gen.Generate(av, username, msg)
	if err != nil {
		return nil, errors.Unexpected("generate image: " + err.Error())
	}

	return bytes.NewReader(buf), err
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
