package commands

import (
	"bytes"
	"encoding/base64"
	"io"
	"log/slog"
	"mime"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/internal/manager"
	"github.com/zanz1n/duvua-bot/internal/utils"
)

var cloneCommandData = discordgo.ApplicationCommand{
	Name:        "clone",
	Description: "Clone um usuário e faça o clone enviar uma mensagem",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Create a user clone and make it send some message",
	},
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "O usuário que deseja clonar",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "The user you want to clone",
			},
			Required: true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "message",
			Description: "A mensagem que o clone irá mandar",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "The message's the clone will send",
			},
			Required: true,
		},
	},
}

func NewCloneCommand() manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: false,
		},
		Data:       &cloneCommandData,
		Category:   manager.CommandCategoryFun,
		NeedsDefer: false,
		Handler:    &CloneCommand{},
	}
}

type CloneCommand struct {
}

func (c *CloneCommand) getBase64Avatar(hs *http.Client, guildId string, member *discordgo.Member) (string, error) {
	var url string
	if member.Avatar == "" {
		url = discordgo.EndpointUserAvatar(member.User.ID, member.User.Avatar)
	} else {
		url = discordgo.EndpointGuildMemberAvatar(guildId, member.User.ID, member.Avatar)
	}

	res, err := hs.Get(url)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", errors.Unexpected("unexpected status code")
	} else if mt, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type")); mt != "image/png" {
		return "", errors.Unexpectedf("unexpected image mime type `%s`", mt)
	}

	buf := bytes.NewBuffer([]byte{})

	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return "", err
	}

	base64d := base64.StdEncoding.EncodeToString(buf.Bytes())

	return "data:image/png;base64," + base64d, nil
}

func (c *CloneCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Member == nil {
		return errors.New("esse comando só pode ser utilizado dentro de um servidor")
	}

	if i.Type != discordgo.InteractionApplicationCommand &&
		i.Type != discordgo.InteractionApplicationCommandAutocomplete {
		return errors.New("interação de tipo inesperado")
	}

	data := i.ApplicationCommandData()

	userOpt := utils.GetOption(data.Options, "user")
	if userOpt == nil {
		return errors.New("opção `user` é necessária")
	} else if userOpt.Type != discordgo.ApplicationCommandOptionUser {
		return errors.New("opção `user` precisa ser um usuário válido")
	}
	userId := userOpt.UserValue(nil).ID

	messageOpt := utils.GetOption(data.Options, "message")
	if messageOpt == nil {
		return errors.New("opção `message` é necessária")
	} else if messageOpt.Type != discordgo.ApplicationCommandOptionString {
		return errors.New("opção `user` precisa ser um texto")
	}
	message := messageOpt.StringValue()

	member, err := s.State.Member(i.GuildID, userId)
	if err != nil {
		member, err = s.GuildMember(i.GuildID, userId)
		if err != nil {
			return err
		}
	}

	avatar, err := c.getBase64Avatar(s.Client, i.GuildID, member)
	if err != nil {
		slog.Error(
			"Failed to fetch user avatar on discord cdn",
			"user_id", userId,
			"error", err,
		)
	}

	auditR := discordgo.WithAuditLogReason("Comando /clone")

	webhook, err := s.WebhookCreate(i.ChannelID, member.DisplayName(), avatar, auditR)
	if err != nil {
		slog.Error("Failed to create discord webhook", "error", err)
		return errors.New("não foi possível criar o clone")
	}

	defer func() {
		if err := s.WebhookDelete(webhook.ID, auditR); err != nil {
			slog.Error("Failed to delete discord webhook", "error", err)
		}
	}()

	params := discordgo.WebhookParams{Content: message}

	_, err = s.WebhookExecute(webhook.ID, webhook.Token, false, &params)
	if err != nil {
		slog.Error("Failed to send webhook message", "webhook_id", webhook.ID, "error", err)
		return errors.New("não foi possível enviar a mensagem desejada")
	}

	return s.InteractionRespond(i.Interaction, utils.BasicResponse("Clone criado!"))
}
