package modcmds

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/manager"
	"github.com/zanz1n/duvua/internal/utils"
)

var clearCommandData = discordgo.ApplicationCommand{
	Name:        "clear",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Limpa um número de mensagens no chat",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Clears a certain number of messages from chat",
	},
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "amount",
			Required:    true,
			Description: "A quantidade de mensagens que deseja limpar (max 100)",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "The amount of messages to clear (max 100)",
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "Filtra para excluir mensagens apenas de um usuário",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Deletes messages only if they were sent by this user",
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Name:        "skip_bots",
			Description: "Pula mensagens enviadas por bots",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Skips messages sent by bots",
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionChannel,
			Name:        "channel",
			Description: "O canal onde as mensagens devem ser excluídas (por padrão este canal)",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "The channel where the messages should be deleted (by default this channel)",
			},
		},
	},
}

func NewClearCommand() *manager.Command {
	return &manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: false,
		},
		Data:     &clearCommandData,
		Category: manager.CommandCategoryModeration,
		Handler:  &ClearCommand{},
	}
}

type ClearCommand struct {
}

func (c *ClearCommand) deleteMsgs(
	s *discordgo.Session,
	i *manager.InteractionCreate,
	amount int,
	channel string,
	user *string,
	skipBots bool,
) error {
	messages, err := s.ChannelMessages(channel, 100, "", "", "")
	if err != nil {
		slog.Error("Failed to fetch channel messages", "channel", channel)
		return err
	}

	slices.SortFunc(messages, func(a *discordgo.Message, b *discordgo.Message) int {
		aMillis := a.Timestamp.UnixMilli()
		bMillis := b.Timestamp.UnixMilli()

		if aMillis > bMillis {
			return 1
		} else if bMillis > aMillis {
			return -1
		} else {
			return 0
		}
	})

	interactionMsg, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		return err
	}

	deleteMsgs := []string{}

	for _, msg := range messages {
		if user != nil && msg.Author.ID != *user {
			continue
		} else if skipBots && msg.Author.Bot {
			continue
		} else if interactionMsg != nil && interactionMsg.ID == msg.ID {
			continue
		}
		if tt, err := discordgo.SnowflakeTimestamp(msg.ID); err == nil {
			if time.Since(tt) > 13*24*time.Hour {
				continue
			}
		}

		deleteMsgs = append(deleteMsgs, msg.ID)
	}

	if len(deleteMsgs) > amount {
		startPos := len(deleteMsgs) - amount
		deleteMsgs = deleteMsgs[startPos:]
	}

	err = s.ChannelMessagesBulkDelete(channel, deleteMsgs)
	if err != nil {
		slog.Error(
			"Failed to bulk delete messages",
			"channel", channel,
			"ammount", fmt.Sprintf("%v/%v", len(deleteMsgs), amount),
			"error", err,
		)

		return errors.New("não foi possível excluir as mensagens")
	}
	slog.Info(
		"Bulk-deleted messages from channel",
		"channel", channel,
		"ammount", fmt.Sprintf("%v/%v", len(deleteMsgs), amount),
	)

	return i.Replyf(s,
		"%v/%v mensagens excluídas do canal <#%s>",
		len(deleteMsgs), amount, channel,
	)
}

func (c *ClearCommand) Handle(s *discordgo.Session, i *manager.InteractionCreate) error {
	const MaxDeleteLimit int64 = 100

	if i.Member == nil {
		return errors.New("esse comando só pode ser utilizado dentro de um servidor")
	}

	hasPerm := utils.HasPerm(i.Member.Permissions, discordgo.PermissionManageMessages)
	if !hasPerm {
		return errors.New("você não tem permissão para usar esse comando")
	}

	amount, err := i.GetIntegerOption("amount", true)
	if err != nil {
		return err
	}

	if amount > MaxDeleteLimit || amount < 1 {
		return errors.New("opção `amount` precisa ser um número entre 1 e 100")
	}

	var userFilter *string = nil
	userId, err := i.GetUserOption("user", false)
	if err != nil {
		return err
	} else if userId != "" {
		userFilter = &userId
	}

	skipBots, err := i.GetBooleanOption("skip_bots", false)
	if err != nil {
		return err
	}

	channel, err := i.GetChannelOption("channel", false)
	if err != nil {
		return err
	}

	if channel != "" {
		ch, err := s.State.Channel(channel)
		if err != nil {
			if err != discordgo.ErrStateNotFound {
				return err
			}

			ch, err = s.Channel(channel)
			if err != nil {
				return err
			}
		}

		if ch.GuildID == "" {
			return errors.New("não foi possível verificar o canal fornecido")
		}
		if ch.Type != discordgo.ChannelTypeGuildText {
			return errors.New("opção `channel` precisa ser um canal de texto válido")
		}
	} else {
		channel = i.ChannelID
	}

	if err = i.DeferReply(s, false); err != nil {
		return err
	}

	return c.deleteMsgs(s, i, int(amount), channel, userFilter, skipBots)
}
