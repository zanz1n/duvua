package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua-bot/internal/errors"
	"github.com/zanz1n/duvua-bot/internal/events"
	"github.com/zanz1n/duvua-bot/internal/manager"
	"github.com/zanz1n/duvua-bot/internal/utils"
	"github.com/zanz1n/duvua-bot/internal/welcome"
)

var welcomeCommandData = discordgo.ApplicationCommand{
	Name:        "welcome",
	Description: "Comandos para configurar a funcionalidade de boas vindas",
	DescriptionLocalizations: &map[discordgo.Locale]string{
		discordgo.EnglishUS: "Commands to configure welcome functionality",
	},
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "enable",
			Description: "Habilita mensagens de boas vindas no servidor",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Enables the welcome message on the server",
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "disable",
			Description: "Disabilita mensagens de boas vindas no servidor",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Disables the welcome message on the server",
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "set-channel",
			Description: "Atualiza o canal de texto no qual as mensagens de boas vindas são enviadas",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Updates the welcome messages text channel",
			},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "O canal de texto",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.EnglishUS: "The text channel",
					},
					Required: true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "set-message",
			Description: "Atualiza a mensagem de boas vindas que é enviada",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Updates the welcome message that will be sent",
			},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "O tipo de mensagem que é enviada",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.EnglishUS: "The type of the message that will be sent",
					},
					Required: true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name: welcome.WelcomeTypeMessage.StringPtBr(),
							NameLocalizations: map[discordgo.Locale]string{
								discordgo.EnglishUS: "Message",
							},
							Value: string(welcome.WelcomeTypeMessage),
						},
						{
							Name: welcome.WelcomeTypeImage.StringPtBr(),
							NameLocalizations: map[discordgo.Locale]string{
								discordgo.EnglishUS: "Image",
							},
							Value: string(welcome.WelcomeTypeImage),
						},
						{
							Name: welcome.WelcomeTypeEmbed.StringPtBr(),
							NameLocalizations: map[discordgo.Locale]string{
								discordgo.EnglishUS: "Embed",
							},
							Value: string(welcome.WelcomeTypeEmbed),
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "Placeholders: {{USER}} (o novo membro), {{GUILD}} (nome do servidor) podem ser usados",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.EnglishUS: "Placeholders: {{USER}} (the new member) " +
							"{{GUILD}} (the name of server) can be used",
					},
					Required: true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "test",
			Description: "Testa a mensagem de boas vindas do servidor",
			DescriptionLocalizations: map[discordgo.Locale]string{
				discordgo.EnglishUS: "Tests the welcome messsage",
			},
		},
	},
}

const (
	welcomeComplementDisabled = ", mas as mensagens de boas vindas estão desabilitadas.\n" +
		"Use `/welcome enable` para habilitá-las."

	welcomeComplementNoChannel = ", mas nenhum canal de texto está definido.\n" +
		"Use `/welcome set-channel` para definir o canal de texto."

	welcomeComplementNoChannelDisabled = ", mas nenhum canal de texto está definido e " +
		"as mensagens de boas vindas estão desabilitadas.\n" +
		"Use `/welcome enable` para habilitá-las e " +
		"`/welcome set-channel` para definir o canal de texto."
)

func NewWelcomeCommand(r welcome.WelcomeRepository, ev *events.MemberAddEvent) manager.Command {
	return manager.Command{
		Accepts: manager.CommandAccept{
			Slash:  true,
			Button: false,
		},
		Data:       &welcomeCommandData,
		Category:   manager.CommandCategoryConfig,
		NeedsDefer: false,
		Handler: &WelcomeCommand{
			r:   r,
			evt: ev,
		},
	}
}

type WelcomeCommand struct {
	r   welcome.WelcomeRepository
	evt *events.MemberAddEvent
}

func (c *WelcomeCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Member == nil || i.GuildID == "" {
		return errors.New("esse comando só pode ser utilizado dentro de um servidor")
	} else if i.Type != discordgo.InteractionApplicationCommand &&
		i.Type != discordgo.InteractionApplicationCommandAutocomplete {
		return errors.New("interação de tipo inesperado")
	}

	if !utils.HasPerm(i.Member.Permissions, discordgo.PermissionAdministrator) {
		return errors.New("você não tem permissão para usar esse comando")
	}

	data := i.ApplicationCommandData()

	subCommand := utils.GetSubCommand(data.Options)
	if subCommand == nil {
		return errors.New("opção `sub-command` é necessária")
	}

	var (
		response string
		err      error
	)
	switch subCommand.Name {
	case "enable":
		response, err = c.handleSetEnabled(i.GuildID, true)
	case "disable":
		response, err = c.handleSetEnabled(i.GuildID, false)
	case "set-channel":
		channelOpt := utils.GetOption(subCommand.Options, "channel")
		if channelOpt == nil {
			return errors.New("opção `channel` é necessária")
		} else if channelOpt.Type != discordgo.ApplicationCommandOptionChannel {
			return errors.New("opção `channel` precisa ser um canal de texto válido")
		}

		channel := channelOpt.ChannelValue(s)
		if channel.Type != discordgo.ChannelTypeGuildText {
			return errors.New("opção `channel` precisa ser um canal de texto válido")
		}

		response, err = c.handleSetChannel(i.GuildID, channel.ID)
	case "set-message":
		typeOpt := utils.GetOption(subCommand.Options, "type")
		if typeOpt == nil {
			return errors.New("opção `type` é necessária")
		} else if typeOpt.Type != discordgo.ApplicationCommandOptionString {
			return errors.New("opção `type` precisa ser (Mensagem, Imagem ou Embed)")
		}

		kind, ok := welcome.WelcomeTypeFromString(typeOpt.StringValue())
		if !ok {
			return errors.New("opção `type` precisa ser (Mensagem, Imagem ou Embed)")
		}

		messageOpt := utils.GetOption(subCommand.Options, "message")
		if messageOpt == nil {
			return errors.New("opção `message` é necessária")
		} else if messageOpt.Type != discordgo.ApplicationCommandOptionString {
			return errors.New("opção `message` precisa ser um texto")
		}

		message := messageOpt.StringValue()

		response, err = c.handleSetMessage(i.GuildID, kind, message)
	case "test":
		i.Member.GuildID = i.GuildID
		response, err = c.handleTest(s, i.Member)
	default:
		return errors.New("opção `sub-command` inválida")
	}

	if err != nil {
		return err
	}
	return s.InteractionRespond(i.Interaction, utils.BasicResponse(response))
}

func (c *WelcomeCommand) handleSetEnabled(guildId string, enabled bool) (string, error) {
	wc, err := c.r.GetById(guildId)
	if err != nil {
		return "", err
	}

	if wc != nil {
		if err = c.r.UpdateEnabled(guildId, enabled); err != nil {
			return "", err
		}
	} else {
		wc, err = c.r.Create(guildId, enabled, nil, nil, nil)
		if err != nil {
			return "", err
		}
	}

	if enabled {
		msgComp := ""
		if wc.ChannelId == nil {
			msgComp = welcomeComplementNoChannel
		}

		if wc.Enabled {
			return "**Mensagem de boas vindas já estava habilitada**" + msgComp, nil
		}
		return "**Mensagem de boas vindas habilitada**" + msgComp, nil
	} else {
		if !wc.Enabled {
			return "**Mensagem de boas vindas já estava desabilitada**", nil
		}
		return "**Mensagem de boas vindas desabilitada**", nil
	}
}

func (c *WelcomeCommand) handleSetChannel(guildId string, channelId string) (string, error) {
	wc, err := c.r.GetById(guildId)
	if err != nil {
		return "", err
	}

	if wc != nil {
		if err = c.r.UpdateChannelId(guildId, channelId); err != nil {
			return "", err
		}
	} else {
		wc, err = c.r.Create(guildId, welcome.DefaultEnabled, &channelId, nil, nil)
		if err != nil {
			return "", err
		}
	}

	msgComp := ""
	if !wc.Enabled {
		msgComp = welcomeComplementDisabled
	}

	return fmt.Sprintf("**Canal de texto alterado para <#%s>**"+msgComp, channelId), nil
}

func (c *WelcomeCommand) handleSetMessage(
	guildId string,
	kind welcome.WelcomeType,
	message string,
) (string, error) {
	wc, err := c.r.GetById(guildId)
	if err != nil {
		return "", err
	}

	if wc != nil {
		if err = c.r.UpdateMessage(guildId, message, kind); err != nil {
			return "", err
		}
	} else {
		wc, err = c.r.Create(guildId, welcome.DefaultEnabled, nil, &message, &kind)
		if err != nil {
			return "", err
		}
	}

	msgComp := ""
	switch {
	case wc.ChannelId == nil && !wc.Enabled:
		msgComp = welcomeComplementNoChannelDisabled
	case wc.ChannelId == nil:
		msgComp = welcomeComplementNoChannel
	case !wc.Enabled:
		msgComp = welcomeComplementDisabled
	}

	return fmt.Sprintf(
		"**Mensagem alterada pra** `%s` **com tipo** `%s`"+msgComp,
		message, kind.StringPtBr(),
	), nil
}

func (c *WelcomeCommand) handleTest(
	s *discordgo.Session,
	member *discordgo.Member,
) (string, error) {
	err := c.evt.Trigger(s, member)
	if err != nil {
		return "", err
	}

	return "Mensagem enviada no canal de texto", nil
}
