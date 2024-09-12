package ticketcmds

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/internal/manager"
	"github.com/zanz1n/duvua/internal/ticket"
)

var ticketCommandData = discordgo.ApplicationCommand{
	Name:        "ticket",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Comandos de tickets",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Ticket commands",
	},
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "create",
			Description: "Cria um ticket",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Creates a ticket",
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "delete",
			Description: "Exclui seus tickets",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Deletes members tickets",
			},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "by-id",
					Description: "Exclui um ticket pelo id",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.EnglishUS: "Deletes a ticket by it's id",
					},
					Options: []*discordgo.ApplicationCommandOption{{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "id",
						Description: "O id do ticket que deseja excluir",
						DescriptionLocalizations: map[discordgo.Locale]string{
							discordgo.EnglishUS: "The id of the ticket you want to delete",
						},
						Required: true,
					}},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "all",
					Description: "Exclui todos os seus tickets nesse servidor",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.EnglishUS: "Deletes all of your tickets in this server",
					},
				},
			},
		},
	},
}

func NewTicketCommand(
	r ticket.TicketRepository,
	configR ticket.TicketConfigRepository,
) manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: true,
		},
		Data:     &ticketCommandData,
		Category: manager.CommandCategoryTicket,
		Handler:  &TicketCommand{r: r, configR: configR},
	}
}

type TicketCommand struct {
	r       ticket.TicketRepository
	configR ticket.TicketConfigRepository
}

func (c *TicketCommand) Handle(s *discordgo.Session, i *manager.InteractionCreate) error {
	if i.Member == nil || i.GuildID == "" {
		return errors.New("esse comando só pode ser utilizado dentro de um servidor")
	}

	if i.Type == discordgo.InteractionMessageComponent {
		data := i.MessageComponentData()

		ids := strings.Split(data.CustomID, "/")
		if 2 > len(ids) {
			return errors.New("interação inválida")
		}

		switch ids[1] {
		case "delete":
			if len(ids) != 3 {
				return errors.New("interação inválida")
			}
			return c.handleDeleteById(s, i, ids[2])

		case "create":
			return c.handleCreate(s, i)
		}

		return errors.New("interação inválida")
	}

	subCommandGroup := i.GetSubCommandGroup()
	subCommand, err := i.GetSubCommand()
	if err != nil {
		return err
	}

	if subCommandGroup != nil {
		switch subCommandGroup.Name {
		case "delete":
			switch subCommand.Name {
			case "by-id":
				id, err := i.GetStringOption("id", true)
				if err != nil {
					return err
				}

				return c.handleDeleteById(s, i, id)

			case "all":
				return c.handleDeleteAll(s, i)

			default:
				return errors.New("opção `sub-command` inválida")
			}
		default:
			return errors.New("opção `sub-command-group` inválida")
		}
	} else {
		switch subCommand.Name {
		case "create":
			return c.handleCreate(s, i)
		default:
			return errors.New("opção `sub-command` inválida")
		}
	}
}

func (c *TicketCommand) handleCreate(s *discordgo.Session, i *manager.InteractionCreate) error {
	cfg, err := c.configR.GetByGuildId(i.GuildID)
	if err != nil {
		return err
	}

	if cfg == nil || !cfg.Enabled {
		return errors.New("os tickets estão desabilitados nesse servidor")
	}

	if !cfg.AllowMultiple {
		ts, err := c.r.GetByMember(i.GuildID, i.Member.User.ID)
		if err != nil {
			return err
		}

		if len(ts) > 0 {
			comp := "um ticket criado"
			if len(ts) > 1 {
				comp = fmt.Sprintf("%d tickets criados", len(ts))
			}
			msg := "Esse servidor permite apenas a criação de um ticket por vez, " +
				"mas **você já tem " + comp + ".**\n" +
				"Dica: Use `/ticket delete all` para excluir seus tickets."

			return i.ReplyEphemeralf(s, msg)
		}
	}

	slug := c.r.GenerateSlug()

	createData := ticketCreateData(slug, cfg, i.Interaction)
	ch, err := s.GuildChannelCreateComplex(
		i.GuildID,
		createData,
		discordgo.WithAuditLogReason(fmt.Sprintf("Ticket %s", slug)),
	)
	if err != nil {
		return err
	}

	t, err := c.r.Create(slug, ch.ID, i.Member.User.ID, i.GuildID)
	if err != nil {
		if _, e := s.ChannelDelete(ch.ID); e != nil {
			slog.Error(
				"Failed to delete channel after ticket create failed",
				"channel_id", ch.ID,
			)
		} else {
			slog.Warn(
				"Deleted channel after ticket create failed",
				"channel_id", ch.ID,
			)
		}
		return err
	}

	go c.sendTicketCreateMessage(s, t)

	ticketUrl := fmt.Sprintf("https://discord.com/channels/%s/%s", t.GuildId, t.ChannelId)

	return i.ReplyEphemeral(s, &manager.InteractionResponse{
		Content: fmt.Sprintf("Seu ticket foi criado, <@%s>", i.Member.User.ID),
		Components: []discordgo.MessageComponent{discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{discordgo.Button{
				Label: "Ir",
				Style: discordgo.LinkButton,
				Emoji: emoji("🚀"),
				URL:   ticketUrl,
			}},
		}},
	})
}

func (c *TicketCommand) sendTicketCreateMessage(s *discordgo.Session, t *ticket.Ticket) {
	msg := fmt.Sprintf(
		"ID: `%s`\nO ticket foi criado nesse canal de texto.\n"+
			"Para excluir, use `/ticket delete by-id id: %s` ou "+
			"clique no botão abaixo, que fará a mesma coisa.",
		t.Slug, t.Slug,
	)
	_, err := s.ChannelMessageSendComplex(t.ChannelId, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Title:       "Ticket criado!",
			Description: msg,
		}},
		Components: []discordgo.MessageComponent{discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{discordgo.Button{
				Label:    "Excluir ticket",
				Style:    discordgo.DangerButton,
				Emoji:    emoji("❌"),
				CustomID: fmt.Sprintf("ticket/delete/%s", t.Slug),
			}},
		}},
	})
	if err != nil {
		slog.Error(
			"Failed to send ticket create message",
			"channel_id", t.ChannelId,
			"error", err,
		)
	}
}

func (c *TicketCommand) handleDeleteById(
	s *discordgo.Session,
	i *manager.InteractionCreate,
	id string,
) error {
	t, err := c.r.GetBySlug(id)
	if err != nil {
		return err
	} else if t == nil {
		return errors.Newf("nenhum ticket `%s` encontrado", id)
	}

	if t.UserId != i.Member.User.ID {
		return errors.Newf("o ticket `%s` não é seu", id)
	}

	t, err = c.r.DeleteBySlug(id)
	if err != nil {
		return err
	} else if t == nil {
		return errors.Newf("nenhum ticket `%s` encontrado", id)
	}

	if _, err = s.ChannelDelete(t.ChannelId); err != nil {
		slog.Warn("Failed to delete ticket channel", "slug", t.Slug)
	}

	// TODO: Dump ticket logs

	if t.ChannelId != i.ChannelID {
		return i.Replyf(s, "Ticket `%s` exluído com sucesso", id)
	}
	return nil
}

func (c *TicketCommand) handleDeleteAll(s *discordgo.Session, i *manager.InteractionCreate) error {
	ts, err := c.r.DeleteByMember(i.GuildID, i.Member.User.ID)
	if err != nil {
		return err
	} else if len(ts) == 0 {
		return errors.New("nenhum ticket encontrado")
	}

	if err := i.DeferReply(s, false); err != nil {
		return err
	}

	containsChannel := false
	for _, t := range ts {
		if t.ChannelId == i.ChannelID {
			containsChannel = true
		}
		if _, err = s.ChannelDelete(t.ChannelId); err != nil {
			slog.Warn(
				"Failed to delete ticket channel",
				"guild_id", i.GuildID,
				"channel_id", t.ChannelId,
			)
		}
	}

	// TODO: Dump ticket logs

	if !containsChannel {
		return i.Replyf(s, "%d tickets excluídos com sucesso", len(ts))
	}
	return nil
}

func ticketCreateData(
	slug string,
	cfg *ticket.TicketConfig,
	i *discordgo.Interaction,
) discordgo.GuildChannelCreateData {
	username := i.Member.User.GlobalName
	if username == "" {
		username = i.Member.User.Username
	}

	var name string
	if cfg.ChannelName != "" {
		name = cfg.ChannelName
	} else if cfg.ChannelCategoryId != "" {
		name = ticket.DefaultConfigCategorizedChannelName
	} else if ticket.TicketSlugLength+len(username) > 12 {
		name = ticket.DefaultConfigShortChannelName
	} else {
		name = ticket.DefaultConfigChannelName
	}

	name = strings.Replace(name, "{{USER}}", username, 1)
	name = strings.Replace(name, "{{ID}}", slug, 1)

	return discordgo.GuildChannelCreateData{
		Name:     name,
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: cfg.ChannelCategoryId,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:   i.Member.User.ID,
				Type: discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionViewChannel |
					discordgo.PermissionSendMessages |
					discordgo.PermissionSendMessagesInThreads |
					discordgo.PermissionEmbedLinks |
					discordgo.PermissionAttachFiles |
					discordgo.PermissionAddReactions |
					discordgo.PermissionUseExternalEmojis |
					discordgo.PermissionUseExternalStickers |
					discordgo.PermissionReadMessageHistory,
				Deny: discordgo.PermissionCreatePrivateThreads |
					discordgo.PermissionCreatePublicThreads |
					discordgo.PermissionMentionEveryone |
					// discordgo.PermissionManageMessages |
					// discordgo.PermissionManageThreads |
					discordgo.PermissionSendTTSMessages,
			},
			{
				ID:   i.GuildID,
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: discordgo.PermissionViewChannel |
					discordgo.PermissionSendMessages |
					discordgo.PermissionReadMessageHistory,
			},
		},
	}
}
