package ticketcmds

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/internal/manager"
	"github.com/zanz1n/duvua-bot/internal/ticket"
	"github.com/zanz1n/duvua-bot/internal/utils"
)

var ticketadminCommandData = discordgo.ApplicationCommand{
	Name:        "ticketadmin",
	Type:        discordgo.ChatApplicationCommand,
	Description: "Comandos administrativos de tickets",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Ticket administrative commands",
	},
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "enable",
			Description: "Habilita a funcionalidade de tickets no servidor",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Enables the ticket functionality on the server",
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "disable",
			Description: "Desabilita a funcionalidade de tickets no servidor",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Disables the ticket functionality on the server",
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "allow-multiple",
			Description: "Permite com que os membros criem mais de um ticket no servidor",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Allows members to create more than one ticket on the server",
			},
			Options: []*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "allow",
				Description: "Permitir?",
				DescriptionLocalizations: map[discordgo.Locale]string{
					discordgo.EnglishUS: "Allow it?",
				},
				Required: true,
			}},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "set-category",
			Description: "Define a categoria na qual o canal dos tickets serão criados",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Defines the category in which the ticket channels will be created",
			},
			Options: []*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "A categoria de canais",
				DescriptionLocalizations: map[discordgo.Locale]string{
					discordgo.EnglishUS: "The channel category",
				},
			}},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "add-permanent",
			Description: "Envia uma mensagem com um botão para criar tickets",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Posts a message with a button to create tickets",
			},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "O canal para enviar a mensagem",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.EnglishUS: "The channel where the message will be sent",
					},
					Required: true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "A mensagem que será enviada",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.EnglishUS: "The message that will be sent",
					},
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
			Name:        "delete",
			Description: "Exclui tickets de membros",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Deletes members tickets",
			},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "by-id",
					Description: "Exclui um ticket pelo seu id",
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
					Description: "Exclui todos os tickets de um membro",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.EnglishUS: "Deletes all the tickets of a member",
					},
					Options: []*discordgo.ApplicationCommandOption{{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "O membro cujos tickets serão excluídos",
						DescriptionLocalizations: map[discordgo.Locale]string{
							discordgo.EnglishUS: "The member whose tickets will be deleted",
						},
						Required: true,
					}},
				},
			},
		},
	},
}

func NewTicketAdminCommand(
	r ticket.TicketRepository,
	configR ticket.TicketConfigRepository,
) manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: false,
		},
		Data:     &ticketadminCommandData,
		Category: manager.CommandCategoryConfig,
		Handler:  &TicketAdminCommand{r: r, configR: configR},
	}
}

type TicketAdminCommand struct {
	r       ticket.TicketRepository
	configR ticket.TicketConfigRepository
}

func (c *TicketAdminCommand) Handle(s *discordgo.Session, i *manager.InteractionCreate) error {
	if i.Member == nil || i.GuildID == "" {
		return errors.New("esse comando só pode ser utilizado dentro de um servidor")
	}
	if !utils.HasPerm(i.Member.Permissions, discordgo.PermissionAdministrator) {
		return errors.New("você não tem permissão para usar esse comando")
	}

	subCommandGroup := i.GetSubCommandGroup()
	subCommand, err := i.GetSubCommand()
	if err != nil {
		return err
	}

	var msg string
	if subCommandGroup != nil {
		switch subCommandGroup.Name {
		case "delete":
			switch subCommand.Name {
			case "by-id":
				id, e := i.GetStringOption("id", true)
				if e != nil {
					return e
				}
				return c.handleDeleteById(s, i, id)

			case "all":
				userId, e := i.GetUserOption("user", true)
				if e != nil {
					return e
				}

				if e = i.DeferReply(s, false); e != nil {
					return e
				}
				return c.handleDeleteAll(s, i, userId)

			default:
				return errors.New("opção `sub-command` inválida")
			}
		default:
			return errors.New("opção `sub-command-group` inválida")
		}
	} else {
		switch subCommand.Name {
		case "enable":
			msg, err = c.handleEnable(i.GuildID, true)

		case "disable":
			msg, err = c.handleEnable(i.GuildID, false)

		case "allow-multiple":
			allow, e := i.GetBooleanOption("allow", true)
			if e != nil {
				return e
			}
			msg, err = c.handleAllowMultiple(i.GuildID, allow)

		case "set-category":
			channelId, e := i.GetChannelOption("channel", false)
			if e != nil {
				return e
			}

			if channelId != "" {
				e = checkChannel(s, channelId, discordgo.ChannelTypeGuildCategory)
				if e != nil {
					return e
				}
			}
			msg, err = c.handleSetCategory(i.GuildID, channelId)

		case "add-permanent":
			channelId, e := i.GetChannelOption("channel", true)
			if e != nil {
				return e
			}
			e = checkChannel(s, channelId, discordgo.ChannelTypeGuildText)
			if e != nil {
				return e
			}

			message, e := i.GetStringOption("message", false)
			if e != nil {
				return e
			}
			if message == "" {
				message = "Clique no botão para criar o ticket"
			} else {
				message += fmt.Sprintf("\n- Mensagem por <@%s>", i.Member.User.ID)
			}

			msg, err = c.handleAddPermanent(s, channelId, message)

		default:
			return errors.New("opção `sub-command` inválida")
		}
	}

	if err != nil {
		return err
	}

	return i.Replyf(s, msg)
}

func (c *TicketAdminCommand) handleDeleteById(
	s *discordgo.Session,
	i *manager.InteractionCreate,
	id string,
) error {
	t, err := c.r.DeleteBySlug(id)
	if err != nil {
		return err
	}

	if t == nil {
		return errors.Newf("nenhum ticket `%s` encontrado", id)
	}

	comp := ""
	if _, err = s.ChannelDelete(t.ChannelId); err != nil {
		slog.Warn("Failed to delete ticket channel", "slug", t.Slug)
		comp += fmt.Sprintf(
			", mas não foi possível excluir o canal de texto <#%s>",
			t.ChannelId,
		)
	}

	// TODO: Dump ticket logs

	if t.ChannelId == i.ChannelID {
		return nil
	}

	return i.Replyf(s,
		"Ticket `%s` de <@%s> excluído com sucesso"+comp,
		t.Slug, t.UserId,
	)
}

func (c *TicketAdminCommand) handleDeleteAll(
	s *discordgo.Session,
	i *manager.InteractionCreate,
	userId string,
) error {
	ts, err := c.r.DeleteByMember(i.GuildID, userId)
	if err != nil {
		return err
	}

	if len(ts) == 0 {
		return errors.Newf("nenhum ticket de <@%s> encontrado", userId)
	}

	errc := []string{}
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
			errc = append(errc, t.Slug)
		}
	}

	// TODO: Dump ticket logs

	if containsChannel {
		return nil
	}

	msg := fmt.Sprintf("%d tickets de <@%s> excluídos com sucesso", len(ts), userId)
	if len(errc) > 0 {
		msg += fmt.Sprintf(
			", mas %d canais de texto não puderam ser excluídos\n"+
				"São dos tickets: ",
			len(errc),
		)

		for i, s := range errc {
			if i != 0 {
				msg += ", "
			}
			msg += s
		}
	}

	return i.Replyf(s, msg)
}

func (c *TicketAdminCommand) handleEnable(guildId string, enabled bool) (string, error) {
	const (
		msgEnabled         = "**Tickets habilitados!**"
		msgAlreadyEnabled  = "**Os tickets já estavam habilitados!**"
		msgDisabled        = "**Tickets desabilitados!**"
		msgAlreadyDisabled = "**Os tickets já estavam desabilitados!**"
	)

	cfg, err := c.configR.GetByGuildId(guildId)
	if err != nil {
		return "", err
	}

	if cfg == nil {
		if enabled == ticket.DefaultConfigEnabled {
			return msgAlreadyDisabled, nil
		}
		_, err = c.configR.Create(ticket.TicketConfigCreateData{
			GuildId: guildId,
			Enabled: &enabled,
		})
		if err != nil {
			return "", err
		}
	} else {
		if cfg.Enabled == enabled {
			if enabled {
				return msgAlreadyEnabled, nil
			} else {
				return msgAlreadyDisabled, nil
			}
		}
		if err = c.configR.UpdateEnabled(guildId, enabled); err != nil {
			return "", err
		}
	}

	if enabled {
		return msgEnabled, nil
	} else {
		return msgDisabled, nil
	}
}

func (c *TicketAdminCommand) handleAllowMultiple(guildId string, allow bool) (string, error) {
	const (
		msgAllowed           = "**Agora um membro pode criar mais de um ticket**"
		msgAlreadyAllowed    = "**Membros já podiam criar mais de um ticket**"
		msgDisallowed        = "**Agora um mebro pode criar apenas um ticket por vez**"
		msgAlreadyDisallowed = "**Membros já podiam criar apenas um ticket por vez**"
		msgCompDisabled      = ", mas os tickets estão desabilitados no servidor.\n" +
			"Use `/ticketadmin enable` para habilitá-los."
	)

	cfg, err := c.configR.GetByGuildId(guildId)
	if err != nil {
		return "", err
	}

	if cfg == nil {
		if allow == ticket.DefaultConfigAllowMultiple {
			return msgAlreadyAllowed + msgCompDisabled, nil
		}

		_, err = c.configR.Create(ticket.TicketConfigCreateData{
			GuildId:       guildId,
			AllowMultiple: &allow,
		})
		if err != nil {
			return "", err
		}

		return msgDisallowed + msgCompDisabled, nil
	}

	if cfg.AllowMultiple == allow {
		var msg string
		if allow {
			msg = msgAlreadyAllowed
		} else {
			msg = msgAlreadyDisallowed
		}
		if !cfg.Enabled {
			msg += msgCompDisabled
		}
		return msg, nil
	}

	if err = c.configR.UpdateAllowMultiple(guildId, allow); err != nil {
		return "", err
	}

	var msg string
	if allow {
		msg = msgAllowed
	} else {
		msg = msgDisallowed
	}
	if !cfg.Enabled {
		msg += msgCompDisabled
	}

	return msg, nil
}

func (c *TicketAdminCommand) handleSetCategory(guildId, channelId string) (string, error) {
	const msgCompDisabled = ", mas os tickets estão desabilitados no servidor.\n" +
		"Use `/ticketadmin enable` para habilitá-los."

	cfg, err := c.configR.GetByGuildId(guildId)
	if err != nil {
		return "", err
	}

	if cfg == nil {
		cfg, err = c.configR.Create(ticket.TicketConfigCreateData{
			GuildId:           guildId,
			ChannelCategoryId: channelId,
		})
		if err != nil {
			return "", err
		}
	} else {
		if err = c.configR.UpdateChannelCategory(guildId, channelId); err != nil {
			return "", err
		}
	}

	var msg string
	if channelId == "" {
		msg = "**Os tickets serão criados sem uma categoria de canais**"
	} else {
		msg = fmt.Sprintf(
			"**Os tickets serão criados na categoria de canais <#%s>**",
			channelId,
		)
	}

	if !cfg.Enabled {
		msg += msgCompDisabled
	}

	return msg, nil
}

func (c *TicketAdminCommand) handleAddPermanent(
	s *discordgo.Session,
	channelId, message string,
) (string, error) {
	msg := discordgo.MessageSend{
		Content: message,
		Components: []discordgo.MessageComponent{discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{discordgo.Button{
				Label:    "Criar ticket",
				Style:    discordgo.PrimaryButton,
				Emoji:    emoji("🎫"),
				CustomID: "ticket/create",
			}},
		}},
	}

	_, err := s.ChannelMessageSendComplex(channelId, &msg)
	if err != nil {
		return "", err
	}

	return "Mensagem enviada com sucesso", nil
}

func checkChannel(s *discordgo.Session, id string, kind discordgo.ChannelType) error {
	ch, e := s.State.Channel(id)
	if e != nil {
		ch, e = s.Channel(id)
	}
	if e != nil || ch.Type != kind {
		switch kind {
		case discordgo.ChannelTypeGuildText:
			return errors.New("opção `channel` precisa ser um canal de texto válido")
		case discordgo.ChannelTypeGuildCategory:
			return errors.New("opção `channel` precisa ser uma categoria de canais válida")
		default:
			return errors.New("opção `channel` precisa ser um canal válido")
		}
	}
	return nil
}

func emoji(name string) *discordgo.ComponentEmoji {
	return &discordgo.ComponentEmoji{Name: name}
}
